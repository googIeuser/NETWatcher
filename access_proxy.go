package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type AccessProxy struct {
	mu           sync.Mutex
	listener     net.Listener
	Port         int
	FragmentSize int
	running      bool
	stopCh       chan struct{}
	wg           sync.WaitGroup
	resolver     *dohResolver
}

type dohResolver struct {
	mu      sync.Mutex
	cache   map[string]dohCacheEntry
	client  *http.Client
	baseURL string
}

type dohCacheEntry struct {
	IP      string
	Expires time.Time
}

type dohJSONResponse struct {
	Status int `json:"Status"`
	Answer []struct {
		Name string `json:"name"`
		Type int    `json:"type"`
		TTL  int    `json:"TTL"`
		Data string `json:"data"`
	} `json:"Answer"`
}

func newDoHResolver() *dohResolver {
	dialer := &net.Dialer{Timeout: 4 * time.Second}
	transport := &http.Transport{
		Proxy:           nil,
		TLSClientConfig: &tls.Config{ServerName: "cloudflare-dns.com", MinVersion: tls.VersionTLS12},
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			_, port, err := net.SplitHostPort(address)
			if err != nil {
				port = "443"
			}
			return dialer.DialContext(ctx, network, net.JoinHostPort("1.1.1.1", port))
		},
		TLSHandshakeTimeout: 4 * time.Second,
	}
	return &dohResolver{
		cache:   map[string]dohCacheEntry{},
		client:  &http.Client{Transport: transport, Timeout: 6 * time.Second},
		baseURL: "https://cloudflare-dns.com/dns-query",
	}
}

func (r *dohResolver) resolve(ctx context.Context, host string) (string, error) {
	host = strings.TrimSuffix(strings.TrimSpace(host), ".")
	if ip := net.ParseIP(host); ip != nil {
		return host, nil
	}
	if host == "" {
		return "", errors.New("empty host")
	}
	r.mu.Lock()
	if cached, ok := r.cache[strings.ToLower(host)]; ok && time.Now().Before(cached.Expires) {
		r.mu.Unlock()
		return cached.IP, nil
	}
	r.mu.Unlock()

	requestURL := r.baseURL + "?name=" + url.QueryEscape(host) + "&type=A"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/dns-json")
	req.Header.Set("User-Agent", "NetWatcher/"+appVersion)
	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DoH returned HTTP %d", resp.StatusCode)
	}
	var decoded dohJSONResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&decoded); err != nil {
		return "", err
	}
	if decoded.Status != 0 {
		return "", fmt.Errorf("DoH status %d", decoded.Status)
	}
	ttl := 60
	var selected string
	for _, answer := range decoded.Answer {
		if answer.Type == 1 && net.ParseIP(answer.Data) != nil {
			selected = answer.Data
			if answer.TTL > 0 {
				ttl = answer.TTL
			}
			break
		}
	}
	if selected == "" {
		return "", fmt.Errorf("no IPv4 answer for %s", host)
	}
	if ttl > 3600 {
		ttl = 3600
	}
	r.mu.Lock()
	r.cache[strings.ToLower(host)] = dohCacheEntry{IP: selected, Expires: time.Now().Add(time.Duration(ttl) * time.Second)}
	r.mu.Unlock()
	return selected, nil
}

func NewAccessProxy(port, fragmentSize int) *AccessProxy {
	if port < 0 || port > 65535 {
		port = 8079
	}
	if fragmentSize < 1 || fragmentSize > 64 {
		fragmentSize = 1
	}
	return &AccessProxy{Port: port, FragmentSize: fragmentSize, resolver: newDoHResolver()}
}

func (p *AccessProxy) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.running {
		return nil
	}
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(p.Port)))
	if err != nil {
		return err
	}
	p.listener = listener
	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		p.Port = tcpAddr.Port
	}
	p.stopCh = make(chan struct{})
	p.running = true
	p.wg.Add(1)
	go p.acceptLoop(listener)
	return nil
}

func (p *AccessProxy) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	listener := p.listener
	p.listener = nil
	if p.stopCh != nil {
		close(p.stopCh)
		p.stopCh = nil
	}
	p.mu.Unlock()
	if listener != nil {
		_ = listener.Close()
	}
	p.wg.Wait()
}

func (p *AccessProxy) Running() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

func (p *AccessProxy) Address() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return net.JoinHostPort("127.0.0.1", strconv.Itoa(p.Port))
}

func (p *AccessProxy) acceptLoop(listener net.Listener) {
	defer p.wg.Done()
	for {
		conn, err := listener.Accept()
		if err != nil {
			p.mu.Lock()
			running := p.running
			p.mu.Unlock()
			if !running {
				return
			}
			continue
		}
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			defer conn.Close()
			_ = p.handleConnection(conn)
		}()
	}
}

func (p *AccessProxy) handleConnection(client net.Conn) error {
	_ = client.SetDeadline(time.Now().Add(20 * time.Second))
	reader := bufio.NewReaderSize(client, 64*1024)
	req, err := http.ReadRequest(reader)
	if err != nil {
		return err
	}
	_ = client.SetDeadline(time.Time{})

	if strings.EqualFold(req.Method, http.MethodConnect) {
		return p.handleConnect(client, reader, req.Host)
	}
	return p.handleHTTP(client, req)
}

func splitAuthority(authority string, defaultPort string) (string, string, error) {
	authority = strings.TrimSpace(authority)
	if authority == "" {
		return "", "", errors.New("empty authority")
	}
	host, port, err := net.SplitHostPort(authority)
	if err == nil {
		return host, port, nil
	}
	if strings.Contains(authority, ":") && !strings.Contains(authority, "]") {
		// Could be host:port without SplitHostPort accepting it due to a malformed port.
		last := strings.LastIndex(authority, ":")
		if last > 0 {
			candidatePort := authority[last+1:]
			if _, parseErr := strconv.Atoi(candidatePort); parseErr == nil {
				return authority[:last], candidatePort, nil
			}
		}
	}
	return strings.Trim(authority, "[]"), defaultPort, nil
}

func (p *AccessProxy) dialTarget(host, port string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	ip, err := p.resolver.resolve(ctx, host)
	if err != nil {
		// A local/IP test target should still work even if DoH is temporarily unavailable.
		addrs, lookupErr := net.DefaultResolver.LookupHost(ctx, host)
		if lookupErr != nil || len(addrs) == 0 {
			return nil, err
		}
		ip = addrs[0]
	}
	dialer := net.Dialer{Timeout: 8 * time.Second, KeepAlive: 30 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, port))
	if tcp, ok := conn.(*net.TCPConn); ok {
		_ = tcp.SetNoDelay(true)
	}
	return conn, err
}

func (p *AccessProxy) handleConnect(client net.Conn, reader *bufio.Reader, authority string) error {
	host, port, err := splitAuthority(authority, "443")
	if err != nil {
		return err
	}
	upstream, err := p.dialTarget(host, port)
	if err != nil {
		_, _ = io.WriteString(client, "HTTP/1.1 502 Bad Gateway\r\nConnection: close\r\n\r\n")
		return err
	}
	defer upstream.Close()
	if _, err := io.WriteString(client, "HTTP/1.1 200 Connection Established\r\nProxy-Agent: NetWatcher\r\n\r\n"); err != nil {
		return err
	}

	errCh := make(chan error, 2)
	go func() {
		_, copyErr := io.Copy(client, upstream)
		errCh <- copyErr
	}()
	go func() {
		copyErr := copyWithInitialFragmentation(upstream, reader, p.FragmentSize)
		errCh <- copyErr
	}()
	first := <-errCh
	return first
}

func (p *AccessProxy) handleHTTP(client net.Conn, req *http.Request) error {
	host := req.URL.Hostname()
	port := req.URL.Port()
	if host == "" {
		host, port, _ = splitAuthority(req.Host, "80")
	}
	if port == "" {
		if strings.EqualFold(req.URL.Scheme, "https") {
			port = "443"
		} else {
			port = "80"
		}
	}
	upstream, err := p.dialTarget(host, port)
	if err != nil {
		_, _ = io.WriteString(client, "HTTP/1.1 502 Bad Gateway\r\nConnection: close\r\n\r\n")
		return err
	}
	defer upstream.Close()

	req.RequestURI = ""
	req.URL.Scheme = ""
	req.URL.Host = ""
	var requestBytes bytes.Buffer
	if err := req.Write(&requestBytes); err != nil {
		return err
	}
	raw := bytes.Replace(requestBytes.Bytes(), []byte("Host:"), []byte("hoSt:"), 1)
	if err := writeFragmented(upstream, raw, p.FragmentSize); err != nil {
		return err
	}
	_, err = io.Copy(client, upstream)
	return err
}

func copyWithInitialFragmentation(dst io.Writer, src io.Reader, fragmentSize int) error {
	buffer := make([]byte, 32*1024)
	first := true
	for {
		n, err := src.Read(buffer)
		if n > 0 {
			var writeErr error
			if first {
				writeErr = writeFragmented(dst, buffer[:n], fragmentSize)
				first = false
			} else {
				_, writeErr = dst.Write(buffer[:n])
			}
			if writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

func writeFragmented(dst io.Writer, data []byte, fragmentSize int) error {
	if fragmentSize < 1 {
		fragmentSize = 1
	}
	initialWindow := len(data)
	if initialWindow > 96 {
		initialWindow = 96
	}
	for offset := 0; offset < initialWindow; offset += fragmentSize {
		end := offset + fragmentSize
		if end > initialWindow {
			end = initialWindow
		}
		if _, err := dst.Write(data[offset:end]); err != nil {
			return err
		}
		if end < initialWindow {
			time.Sleep(2 * time.Millisecond)
		}
	}
	if initialWindow < len(data) {
		_, err := dst.Write(data[initialWindow:])
		return err
	}
	return nil
}
