//go:build windows

package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"netwatcher/wizard"
)

const (
	appName    = "NetWatcher"
	appVersion = "2.1.3"

	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_CAPTION          = 0x00C00000
	WS_SYSMENU          = 0x00080000
	WS_MINIMIZEBOX      = 0x00020000
	WS_CLIPCHILDREN     = 0x02000000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_BORDER           = 0x00800000
	WS_VSCROLL          = 0x00200000
	WS_TABSTOP          = 0x00010000
	WS_DISABLED         = 0x08000000

	WS_EX_DLGMODALFRAME = 0x00000001
	WS_EX_CONTROLPARENT = 0x00010000

	ES_LEFT        = 0x0000
	ES_MULTILINE   = 0x0004
	ES_AUTOVSCROLL = 0x0040
	ES_READONLY    = 0x0800
	ES_AUTOHSCROLL = 0x0080

	BS_PUSHBUTTON    = 0x00000000
	BS_DEFPUSHBUTTON = 0x00000001
	BS_AUTOCHECKBOX  = 0x00000003
	BS_OWNERDRAW     = 0x0000000B

	SS_LEFT   = 0x00000000
	SS_CENTER = 0x00000001

	CBS_DROPDOWN     = 0x0002
	CBS_DROPDOWNLIST = 0x0003
	CBS_AUTOHSCROLL  = 0x0040

	SW_HIDE       = 0
	SW_SHOW       = 5
	SW_SHOWNORMAL = 1

	WM_CREATE         = 0x0001
	WM_DESTROY        = 0x0002
	WM_SIZE           = 0x0005
	WM_PAINT          = 0x000F
	WM_CLOSE          = 0x0010
	WM_ERASEBKGND     = 0x0014
	WM_CTLCOLOREDIT   = 0x0133
	WM_CTLCOLORBTN    = 0x0135
	WM_CTLCOLORSTATIC = 0x0138
	WM_COMMAND        = 0x0111
	WM_DRAWITEM       = 0x002B
	WM_SETICON        = 0x0080
	WM_SETFONT        = 0x0030
	WM_APP            = 0x8000
	WM_USER           = 0x0400

	BN_CLICKED    = 0
	CBN_SELCHANGE = 1

	MB_OK              = 0x00000000
	MB_OKCANCEL        = 0x00000001
	MB_YESNO           = 0x00000004
	MB_ICONINFORMATION = 0x00000040
	MB_ICONQUESTION    = 0x00000020
	MB_ICONERROR       = 0x00000010
	MB_ICONWARNING     = 0x00000030

	IDOK     = 1
	IDCANCEL = 2
	IDYES    = 6
	IDNO     = 7

	COLOR_WINDOW    = 5
	IDC_ARROW       = 32512
	IDI_APPLICATION = 32512
	IMAGE_ICON      = 1
	LR_LOADFROMFILE = 0x0010
	LR_DEFAULTSIZE  = 0x0040
	ICON_SMALL      = 0
	ICON_BIG        = 1

	DEFAULT_GUI_FONT = 17
	TRANSPARENT      = 1
	PS_SOLID         = 0
	NULL_BRUSH       = 5

	FW_NORMAL         = 400
	FW_SEMIBOLD       = 600
	FW_BOLD           = 700
	DEFAULT_CHARSET   = 1
	CLEARTYPE_QUALITY = 5

	BIF_RETURNONLYFSDIRS = 0x0001
	BIF_EDITBOX          = 0x0010
	BIF_NEWDIALOGSTYLE   = 0x0040
	BFFM_INITIALIZED     = 1
	BFFM_SETSELECTIONW   = WM_USER + 103

	CREATE_NO_WINDOW = 0x08000000

	ODS_SELECTED = 0x0001
	ODS_DISABLED = 0x0004
	ODS_FOCUS    = 0x0010

	DT_LEFT         = 0x00000000
	DT_CENTER       = 0x00000001
	DT_RIGHT        = 0x00000002
	DT_VCENTER      = 0x00000004
	DT_SINGLELINE   = 0x00000020
	DT_WORDBREAK    = 0x00000010
	DT_END_ELLIPSIS = 0x00008000

	RDW_INVALIDATE  = 0x0001
	RDW_ERASE       = 0x0004
	RDW_ALLCHILDREN = 0x0080
	RDW_UPDATENOW   = 0x0100

	CB_GETCURSEL    = 0x0147
	CB_ADDSTRING    = 0x0143
	CB_RESETCONTENT = 0x014B
	CB_SETCURSEL    = 0x014E
	BM_GETCHECK     = 0x00F0
	BM_SETCHECK     = 0x00F1
	BST_CHECKED     = 1

	ICC_PROGRESS_CLASS = 0x00000020
	PBM_SETPOS         = WM_USER + 2
	PBM_SETRANGE32     = WM_USER + 6

	ctrlStart      = 1001
	ctrlStop       = 1002
	ctrlReport     = 1003
	ctrlLogs       = 1004
	ctrlAdd        = 1005
	ctrlInterval   = 1006
	ctrlCustom     = 1008
	ctrlStatusText = 1009
	ctrlEventText  = 1010
	ctrlSummary    = 1011
	ctrlSettings   = 1012
	ctrlStats      = 1013
	ctrlExport     = 1014
	ctrlRemove     = 1015

	staticInterval = 2001
	staticCustom   = 2003

	setLanguage       = 5001
	setTheme          = 5010
	setInterval       = 5002
	setTimeout        = 5003
	setLatency        = 5004
	setConfirm        = 5005
	setAutoStart      = 5006
	setAutoMonitor    = 5007
	setCloseToTray    = 5011
	setStartMinimized = 5012
	setSave           = 5008
	setCancel         = 5009
	setLabelLanguage  = 5101
	setLabelTheme     = 5110
	setLabelInterval  = 5102
	setLabelTimeout   = 5103
	setLabelLatency   = 5104
	setLabelConfirm   = 5105
	setInfo           = 5106
	setAutoUpdate     = 5111
	setOutageNotify   = 5112

	instLanguageCombo = wizard.CtrlLanguageCombo
	instPathEdit      = wizard.CtrlPathEdit
	instBrowse        = wizard.CtrlBrowse
	instDesktop       = wizard.CtrlDesktop
	instStartMenu     = wizard.CtrlStartMenu
	instLaunch        = wizard.CtrlLaunch
	instBack          = wizard.CtrlBack
	instNext          = wizard.CtrlNext
	instCancel        = wizard.CtrlCancel
	instTitle         = 6010
	instSubtitle      = 6011
	instSummary       = wizard.CtrlSummary
	instProgress      = wizard.CtrlProgress
	instDetails       = wizard.CtrlDetails
	instPageText      = 6015
	instPathLabel     = 6016
	instOptionsText   = 6017
	instLanguageLabel = 6018
	instAutoStart     = wizard.CtrlAutoStart
	instDetailsLabel  = 6020
	instFinishText    = 6021

	msgRefresh           = WM_APP + 1
	msgInstallerProgress = WM_APP + 2
	msgInstallerDone     = WM_APP + 3
	msgTrayIcon          = WM_APP + 20
	msgSyncTray          = WM_APP + 21

	trayIconID        = 1
	trayOpenID        = 7101
	trayExitID        = 7102
	trayCheckUpdateID = 7103

	NIM_ADD     = 0x00000000
	NIM_MODIFY  = 0x00000001
	NIM_DELETE  = 0x00000002
	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004

	WM_LBUTTONDBLCLK = 0x0203
	WM_RBUTTONUP     = 0x0205
	MF_STRING        = 0x00000000
	MF_SEPARATOR     = 0x00000800
	TPM_RIGHTBUTTON  = 0x0002
	TPM_RETURNCMD    = 0x0100

	maxHistory = 300
)

//go:embed assets/app_icon.ico
var embeddedAppIcon []byte

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")
	uxtheme  = syscall.NewLazyDLL("uxtheme.dll")
	dwmapi   = syscall.NewLazyDLL("dwmapi.dll")

	procRegisterClassExW     = user32.NewProc("RegisterClassExW")
	procCreateWindowExW      = user32.NewProc("CreateWindowExW")
	procDefWindowProcW       = user32.NewProc("DefWindowProcW")
	procShowWindow           = user32.NewProc("ShowWindow")
	procUpdateWindow         = user32.NewProc("UpdateWindow")
	procGetMessageW          = user32.NewProc("GetMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procDestroyWindow        = user32.NewProc("DestroyWindow")
	procMessageBoxW          = user32.NewProc("MessageBoxW")
	procLoadCursorW          = user32.NewProc("LoadCursorW")
	procLoadIconW            = user32.NewProc("LoadIconW")
	procLoadImageW           = user32.NewProc("LoadImageW")
	procDrawTextW            = user32.NewProc("DrawTextW")
	procSetWindowTextW       = user32.NewProc("SetWindowTextW")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procMoveWindow           = user32.NewProc("MoveWindow")
	procSendMessageW         = user32.NewProc("SendMessageW")
	procPostMessageW         = user32.NewProc("PostMessageW")
	procInvalidateRect       = user32.NewProc("InvalidateRect")
	procRedrawWindow         = user32.NewProc("RedrawWindow")
	procBeginPaint           = user32.NewProc("BeginPaint")
	procEndPaint             = user32.NewProc("EndPaint")
	procGetClientRect        = user32.NewProc("GetClientRect")
	procGetWindowRect        = user32.NewProc("GetWindowRect")
	procFillRect             = user32.NewProc("FillRect")
	procEnableWindow         = user32.NewProc("EnableWindow")
	procSetForegroundWindow  = user32.NewProc("SetForegroundWindow")
	procCreatePopupMenu      = user32.NewProc("CreatePopupMenu")
	procAppendMenuW          = user32.NewProc("AppendMenuW")
	procTrackPopupMenu       = user32.NewProc("TrackPopupMenu")
	procDestroyMenu          = user32.NewProc("DestroyMenu")
	procGetCursorPos         = user32.NewProc("GetCursorPos")

	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procCreatePen        = gdi32.NewProc("CreatePen")
	procSelectObject     = gdi32.NewProc("SelectObject")
	procDeleteObject     = gdi32.NewProc("DeleteObject")
	procMoveToEx         = gdi32.NewProc("MoveToEx")
	procLineTo           = gdi32.NewProc("LineTo")
	procSetTextColor     = gdi32.NewProc("SetTextColor")
	procSetBkColor       = gdi32.NewProc("SetBkColor")
	procSetBkMode        = gdi32.NewProc("SetBkMode")
	procTextOutW         = gdi32.NewProc("TextOutW")
	procRectangle        = gdi32.NewProc("Rectangle")
	procRoundRect        = gdi32.NewProc("RoundRect")
	procEllipse          = gdi32.NewProc("Ellipse")
	procArc              = gdi32.NewProc("Arc")
	procGetStockObject   = gdi32.NewProc("GetStockObject")
	procCreateFontW      = gdi32.NewProc("CreateFontW")

	procGetModuleHandleW         = kernel32.NewProc("GetModuleHandleW")
	procGetUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")
	procInitCommonControlsEx     = comctl32.NewProc("InitCommonControlsEx")

	procSHBrowseForFolderW    = shell32.NewProc("SHBrowseForFolderW")
	procShellNotifyIconW      = shell32.NewProc("Shell_NotifyIconW")
	procSHGetPathFromIDListW  = shell32.NewProc("SHGetPathFromIDListW")
	procOleInitialize         = ole32.NewProc("OleInitialize")
	procOleUninitialize       = ole32.NewProc("OleUninitialize")
	procCoTaskMemFree         = ole32.NewProc("CoTaskMemFree")
	procSetWindowTheme        = uxtheme.NewProc("SetWindowTheme")
	procDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
)

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       syscall.Handle
}

type POINT struct{ X, Y int32 }
type RECT struct{ Left, Top, Right, Bottom int32 }
type MSG struct {
	HWnd     syscall.Handle
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       POINT
	LPrivate uint32
}
type PAINTSTRUCT struct {
	Hdc         syscall.Handle
	FErase      int32
	RcPaint     RECT
	FRestore    int32
	FIncUpdate  int32
	RGBReserved [32]byte
}
type INITCOMMONCONTROLSEX struct {
	DwSize uint32
	DwICC  uint32
}

type NOTIFYICONDATA struct {
	CbSize            uint32
	HWnd              syscall.Handle
	UID               uint32
	UFlags            uint32
	UCallbackMessage  uint32
	HIcon             syscall.Handle
	SzTip             [128]uint16
	DwState           uint32
	DwStateMask       uint32
	SzInfo            [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle       [64]uint16
	DwInfoFlags       uint32
	GuidItem          [16]byte
	HBalloonIcon      syscall.Handle
}

type DRAWITEMSTRUCT struct {
	CtlType    uint32
	CtlID      uint32
	ItemID     uint32
	ItemAction uint32
	ItemState  uint32
	HwndItem   syscall.Handle
	HDC        syscall.Handle
	RcItem     RECT
	ItemData   uintptr
}

type BROWSEINFO struct {
	HwndOwner      syscall.Handle
	PidlRoot       uintptr
	PszDisplayName *uint16
	LpszTitle      *uint16
	UlFlags        uint32
	Lpfn           uintptr
	LParam         uintptr
	IImage         int32
}

type Config struct {
	Language            string   `json:"language"`
	Theme               string   `json:"theme"`
	Interval            float64  `json:"interval_seconds"`
	TimeoutMS           int      `json:"timeout_ms"`
	HighLatencyMS       float64  `json:"high_latency_ms"`
	ConfirmCycles       int      `json:"confirm_cycles"`
	AutoStart           bool     `json:"start_with_windows"`
	StartMinimizedTray  bool     `json:"start_minimized_to_notification_area"`
	AutoMonitor         bool     `json:"start_monitoring_automatically"`
	CloseToTray         bool     `json:"keep_running_in_tray_on_close"`
	AutoCheckUpdates    bool     `json:"automatically_check_for_updates"`
	OutageNotifications bool     `json:"show_outage_notifications"`
	FirstRunComplete    bool     `json:"first_run_setup_completed"`
	CustomTargets       []string `json:"custom_targets"`
}

type Target struct {
	Name string
	Host string
	Kind string
}

type PingResult struct {
	Timestamp time.Time
	Target    Target
	Success   bool
	Latency   float64
	Message   string
}

type Sample struct {
	Time    time.Time
	Latency float64
	Success bool
}

type Outage struct {
	Start    time.Time
	End      time.Time
	Category string
	Details  string
}

type App struct {
	hwnd              syscall.Handle
	controls          map[int]syscall.Handle
	mu                sync.RWMutex
	targets           []Target
	latest            map[string]PingResult
	history           map[string][]Sample
	events            []string
	results           []PingResult
	outages           []Outage
	active            *Outage
	pendingState      string
	pendingCount      int
	monitoring        bool
	stopCh            chan struct{}
	logDir            string
	config            Config
	trayAdded         bool
	trayIcon          syscall.Handle
	trayMu            sync.Mutex
	startHidden       bool
	exiting           bool
	notificationQueue []TrayNotification
	pendingUpdateURL  string
}

type SettingsWindow struct {
	hwnd          syscall.Handle
	controls      map[int]syscall.Handle
	parent        *App
	originalLang  string
	originalTheme string
}

type Installer struct {
	hwnd         syscall.Handle
	controls     map[int]syscall.Handle
	pageControls map[wizard.Page][]syscall.Handle
	page         wizard.Page
	mu           sync.Mutex
	language     string
	installDir   string
	installing   bool
	installed    bool
	installErr   string
	installedExe string
	logs         []string
	progress     int
	options      map[int]bool
	titleFont    syscall.Handle
	brandFont    syscall.Handle
	taglineFont  syscall.Handle
	headerFont   syscall.Handle
	normalFont   syscall.Handle
	smallFont    syscall.Handle
}

var globalApp *App
var globalSettings *SettingsWindow
var globalInstaller *Installer
var lightBackgroundBrush syscall.Handle
var lightInputBrush syscall.Handle
var darkBackgroundBrush syscall.Handle
var darkInputBrush syscall.Handle
var wndProcCallback = syscall.NewCallback(windowProc)
var settingsProcCallback = syscall.NewCallback(settingsProc)
var installerProcCallback = syscall.NewCallback(installerProc)
var folderBrowseCallback = syscall.NewCallback(folderBrowseProc)
var latencyRe = regexp.MustCompile(`(?i)(?:time|s.re)?\s*[=<]\s*(\d+(?:[\.,]\d+)?)\s*ms`)

var translations = map[string]map[string]string{
	"tr": {
		"start": "▶ İzlemeyi Başlat", "stop": "■ Durdur", "report": "HTML Raporu", "logs": "Log Klasörü", "settings": "⚙ Ayarlar",
		"interval": "Aralık (sn):", "timeout": "Zaman aşımı (ms):", "custom": "Özel hedef:", "add": "Hedef Ekle", "remove": "Hedefi Sil",
		"target": "Hedef", "address": "Adres", "type": "Tür", "status": "Durum", "latency": "Gecikme", "last": "Son Kontrol",
		"waiting": "Bekliyor", "online": "Çevrimiçi", "failed": "Başarısız", "internet": "İnternet", "local": "Yerel",
		"monitor_running": "İzleme çalışıyor", "monitor_stopped": "İzleme durduruldu", "not_started": "İzleme başlatılmadı",
		"samples": "Oturum örnekleri", "outages": "Kesinti", "total": "Toplam", "active_outage": "Aktif kesinti var",
		"invalid_values": "Aralık en az 0,5 saniye; zaman aşımı en az 200 ms olmalıdır.", "invalid_interval": "Aralık en az 0,5 saniye olmalıdır.", "no_report": "Rapor için henüz ölçüm bulunmuyor.",
		"monitor_started_event": "İzleme başlatıldı.", "monitor_stopped_event": "İzleme durduruldu.",
		"gateway_missing": "Varsayılan ağ geçidi otomatik bulunamadı. İnternet hedefleri yine izlenecek.",
		"target_added":    "Özel hedef eklendi", "target_removed": "Özel hedef silindi", "remove_confirm": "Bu özel hedef silinsin mi?\n\n%s", "select_custom_target": "Silmek için açılır listeden bir özel hedef seçin veya adresini yazın.", "custom_target_not_found": "Bu adres kayıtlı bir özel hedef değil.", "response_ok": "Yanıt alındı", "ping_error": "Zaman aşımı veya erişim hatası",
		"local_error": "Modem/ağ geçidine erişilemiyor", "isp_error": "Modeme erişim var ancak tüm internet hedefleri başarısız.",
		"partial_error": "Bazı internet hedefleri başarısız", "high_latency": "Yüksek gecikme", "connection_normal": "Bağlantı normal.",
		"outage_started": "KESİNTİ BAŞLADI", "outage_ended": "KESİNTİ BİTTİ", "back_normal": "Bağlantı normale döndü.", "state_changed": "Durum değişti.",
		"user_stopped": "İzleme kullanıcı tarafından durduruldu", "app_closed": "Uygulama kapatıldı",
		"state_local": "Yerel ağ/modem", "state_isp": "ISS/internet", "state_degraded": "Kısmi erişim", "state_high": "Yüksek gecikme", "state_online": "Normal",
		"graph_fail":     "Kırmızı çizgi = başarısız ping",
		"settings_title": "NetWatcher Ayarları", "language": "Uygulama dili:", "theme": "Görünüm teması:", "theme_light": "Aydınlık", "theme_dark": "Karanlık", "high_latency_label": "Yüksek gecikme eşiği (ms):",
		"confirm_label": "Kesinti doğrulama ölçümü:", "auto_start": "Windows açıldığında NetWatcher'ı başlat",
		"auto_monitor": "Uygulama açılınca izlemeyi otomatik başlat", "start_minimized_tray": "Windows başladığında bildirim alanına küçültülmüş olarak başlat", "close_to_tray": "Pencere kapatıldığında bildirim alanında çalışmaya devam et", "save": "Kaydet", "cancel": "İptal",
		"settings_info":  "Dil ve tema değişiklikleri anında önizlenir; Kaydet düğmesiyle kalıcı hale gelir.",
		"settings_saved": "Ayarlar kaydedildi.", "settings_invalid": "Ayar değerlerinden biri geçersiz.",
		"report_title": "NetWatcher Bağlantı Raporu", "created": "Oluşturulma", "measurement_range": "Ölçüm aralığı",
		"total_samples": "Toplam ping örneği", "completed_outages": "Tamamlanan kesinti", "total_outage": "Toplam kesinti süresi",
		"target_summary": "Hedef Özeti", "sample": "Örnek", "packet_loss": "Paket kaybı", "average": "Ortalama",
		"outage_events": "Kesinti Olayları", "start_time": "Başlangıç", "end_time": "Bitiş", "duration": "Süre", "class": "Sınıf", "description": "Açıklama",
		"no_outage": "Kesinti kaydı yok.", "print_pdf": "Yazdır / PDF Olarak Kaydet",
		"report_note":    "Ham kayıtlar Belgeler\\NetWatcherLogs klasöründeki günlük CSV dosyalarında bulunur. BTK veya ISS başvurusu için “Yazdır → PDF olarak kaydet” seçeneğini kullanabilirsiniz.",
		"report_created": "HTML raporu oluşturuldu",
	},
	"en": {
		"start": "▶ Start Monitoring", "stop": "■ Stop", "report": "HTML Report", "logs": "Log Folder", "settings": "⚙ Settings",
		"interval": "Interval (sec):", "timeout": "Timeout (ms):", "custom": "Custom target:", "add": "Add Target", "remove": "Remove Target",
		"target": "Target", "address": "Address", "type": "Type", "status": "Status", "latency": "Latency", "last": "Last Check",
		"waiting": "Waiting", "online": "Online", "failed": "Failed", "internet": "Internet", "local": "Local",
		"monitor_running": "Monitoring is running", "monitor_stopped": "Monitoring stopped", "not_started": "Monitoring has not started",
		"samples": "Session samples", "outages": "Outages", "total": "Total", "active_outage": "Active outage",
		"invalid_values": "Interval must be at least 0.5 seconds and timeout at least 200 ms.", "invalid_interval": "Interval must be at least 0.5 seconds.", "no_report": "There are no measurements for a report yet.",
		"monitor_started_event": "Monitoring started.", "monitor_stopped_event": "Monitoring stopped.",
		"gateway_missing": "The default gateway could not be detected automatically. Internet targets will still be monitored.",
		"target_added":    "Custom target added", "target_removed": "Custom target removed", "remove_confirm": "Remove this custom target?\n\n%s", "select_custom_target": "Select a custom target from the drop-down list or enter its address.", "custom_target_not_found": "This address is not a saved custom target.", "response_ok": "Reply received", "ping_error": "Request timed out or destination is unreachable",
		"local_error": "The modem/default gateway is unreachable", "isp_error": "The modem is reachable but all internet targets failed.",
		"partial_error": "Some internet targets failed", "high_latency": "High latency", "connection_normal": "Connection is normal.",
		"outage_started": "OUTAGE STARTED", "outage_ended": "OUTAGE ENDED", "back_normal": "Connection returned to normal.", "state_changed": "Status changed.",
		"user_stopped": "Monitoring was stopped by the user", "app_closed": "Application closed",
		"state_local": "Local network/modem", "state_isp": "ISP/internet", "state_degraded": "Partial access", "state_high": "High latency", "state_online": "Normal",
		"graph_fail":     "Red line = failed ping",
		"settings_title": "NetWatcher Settings", "language": "Application language:", "theme": "Appearance theme:", "theme_light": "Light", "theme_dark": "Dark", "high_latency_label": "High latency threshold (ms):",
		"confirm_label": "Outage confirmation samples:", "auto_start": "Start NetWatcher with Windows",
		"auto_monitor": "Start monitoring automatically when the app opens", "start_minimized_tray": "Start minimized to the notification area when Windows starts", "close_to_tray": "Keep NetWatcher running in the notification area when the window is closed", "save": "Save", "cancel": "Cancel",
		"settings_info":  "Language and theme changes are previewed instantly and become permanent after saving.",
		"settings_saved": "Settings saved.", "settings_invalid": "One or more setting values are invalid.",
		"report_title": "NetWatcher Connection Report", "created": "Created", "measurement_range": "Measurement range",
		"total_samples": "Total ping samples", "completed_outages": "Completed outages", "total_outage": "Total outage duration",
		"target_summary": "Target Summary", "sample": "Samples", "packet_loss": "Packet loss", "average": "Average",
		"outage_events": "Outage Events", "start_time": "Start", "end_time": "End", "duration": "Duration", "class": "Class", "description": "Description",
		"no_outage": "No outage records.", "print_pdf": "Print / Save as PDF",
		"report_note":    "Raw measurements are stored in daily CSV files under Documents\\NetWatcherLogs. For ISP submissions, use Print → Save as PDF.",
		"report_created": "HTML report created",
	},
}

var installerTranslations = map[string]map[string]string{
	"tr": {
		"window": "NetWatcher Kurulum", "back": "< Geri", "next": "İleri >", "install": "Kur", "finish": "Bitir", "cancel": "İptal",
		"welcome_title": "NetWatcher Kurulum Sihirbazına Hoş Geldiniz",
		"welcome_sub":   "Bu sihirbaz NetWatcher'ı bilgisayarınıza kuracaktır. Devam etmek için İleri düğmesine tıklayın.",
		"welcome_body":  "İnternet kesintilerini, paket kaybını ve gecikmeyi kayıt altına alın. Kurulum ve uygulama dilini aşağıdan seçebilirsiniz.",
		"lang_text":     "Kurulum ve uygulama dili:",
		"options_title": "Kurulum Seçeneklerini Seçin", "options_sub": "NetWatcher'ın nereye ve nasıl kurulacağını seçin.",
		"path_label": "Klasör:", "browse": "Gözat...",
		"path_note": "NetWatcher bu klasöre kurulacaktır. Ölçüm kayıtları Belgeler\\NetWatcherLogs klasöründe ayrı tutulur.",
		"desktop":   "Masaüstü kısayolu oluştur", "startmenu": "Başlat menüsü kısayolu oluştur",
		"autostart":     "Windows açıldığında NetWatcher'ı başlat",
		"launch":        "Kurulum tamamlanınca NetWatcher'ı başlat",
		"summary_title": "NetWatcher Kuruluma Hazır", "summary_sub": "Kurulumu başlatmak için Kur düğmesine tıklayın. Seçimleri değiştirmek için Geri'yi kullanın.",
		"installing_title": "NetWatcher Kuruluyor", "installing_sub": "Lütfen dosyalar kopyalanırken bekleyin.",
		"details":      "Kurulum ayrıntıları:",
		"finish_title": "Kurulum Tamamlandı", "finish_sub": "NetWatcher başarıyla bilgisayarınıza kuruldu.",
		"finish_text":  "NetWatcher kullanıma hazırdır. Ayarlarınızı uygulama içindeki Ayarlar bölümünden değiştirebilirsiniz.\r\n\r\nÖlçüm kayıtları Belgeler\\NetWatcherLogs klasöründe saklanacaktır.",
		"summary_path": "Kurulum klasörü", "summary_desktop": "Masaüstü kısayolu", "summary_startmenu": "Başlat menüsü kısayolu", "summary_autostart": "Windows ile başlat", "summary_language": "Dil",
		"yes": "Evet", "no": "Hayır", "confirm_cancel": "Kurulumdan çıkmak istediğinizden emin misiniz?", "invalid_path": "Geçerli bir kurulum klasörü seçin.",
		"folder_dialog":  "NetWatcher için kurulum klasörünü seçin",
		"install_failed": "Kurulum tamamlanamadı", "detail_prepare": "Kurulum hazırlanıyor...", "detail_copy": "Uygulama dosyası kopyalanıyor...",
		"detail_settings": "Başlangıç ayarları kaydediliyor...", "detail_shortcuts": "Kısayollar oluşturuluyor...", "detail_registry": "Windows uygulama kaydı oluşturuluyor...", "detail_done": "Kurulum başarıyla tamamlandı.", "detail_warning": "Uyarı",
	},
	"en": {
		"window": "NetWatcher Setup", "back": "< Back", "next": "Next >", "install": "Install", "finish": "Finish", "cancel": "Cancel",
		"welcome_title": "Welcome to the NetWatcher Setup Wizard",
		"welcome_sub":   "This wizard will install NetWatcher on your computer. Click Next to continue.",
		"welcome_body":  "Record internet outages, packet loss and latency. Choose the setup and application language below.",
		"lang_text":     "Setup and application language:",
		"options_title": "Select Installation Options", "options_sub": "Choose where and how NetWatcher will be installed.",
		"path_label": "Folder:", "browse": "Browse...",
		"path_note": "NetWatcher will be installed in this folder. Measurements are stored separately under Documents\\NetWatcherLogs.",
		"desktop":   "Create a desktop shortcut", "startmenu": "Create a Start menu shortcut",
		"autostart":     "Start NetWatcher when Windows starts",
		"launch":        "Launch NetWatcher when setup finishes",
		"summary_title": "Ready to Install NetWatcher", "summary_sub": "Click Install to begin. Click Back to review or change your selections.",
		"installing_title": "Installing NetWatcher", "installing_sub": "Please wait while files are copied.",
		"details":      "Installation details:",
		"finish_title": "Installation Complete", "finish_sub": "NetWatcher was successfully installed on your computer.",
		"finish_text":  "NetWatcher is ready to use. You can change its settings from the Settings section in the application.\r\n\r\nMeasurements will be stored under Documents\\NetWatcherLogs.",
		"summary_path": "Installation folder", "summary_desktop": "Desktop shortcut", "summary_startmenu": "Start menu shortcut", "summary_autostart": "Start with Windows", "summary_language": "Language",
		"yes": "Yes", "no": "No", "confirm_cancel": "Are you sure you want to exit setup?", "invalid_path": "Select a valid installation folder.",
		"folder_dialog":  "Select the NetWatcher installation folder",
		"install_failed": "Installation could not be completed", "detail_prepare": "Preparing installation...", "detail_copy": "Copying application file...",
		"detail_settings": "Saving initial settings...", "detail_shortcuts": "Creating shortcuts...", "detail_registry": "Registering the application with Windows...", "detail_done": "Installation completed successfully.", "detail_warning": "Warning",
	},
}

func ptr(s string) *uint16     { return syscall.StringToUTF16Ptr(s) }
func loword(v uintptr) uint16  { return uint16(v & 0xffff) }
func hiword(v uintptr) uint16  { return uint16((v >> 16) & 0xffff) }
func rgb(r, g, b byte) uintptr { return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16 }

func normalizeLanguage(_ string) string { return "en" }
func tr(_ string, key string) string {
	if value, ok := translations["en"][key]; ok {
		return value
	}
	return key
}
func it(_ string, key string) string {
	if value, ok := installerTranslations["en"][key]; ok {
		return value
	}
	return key
}

func messageBox(title, text string, flags uintptr) int {
	return messageBoxOwned(0, title, text, flags)
}
func messageBoxOwned(owner syscall.Handle, title, text string, flags uintptr) int {
	r, _, _ := procMessageBoxW.Call(uintptr(owner), uintptr(unsafe.Pointer(ptr(text))), uintptr(unsafe.Pointer(ptr(title))), flags)
	return int(r)
}
func hiddenCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: CREATE_NO_WINDOW}
	return cmd
}

func runHiddenTimeout(timeout time.Duration, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: CREATE_NO_WINDOW}
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return output, fmt.Errorf("%s timed out after %s", name, timeout)
	}
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text != "" {
			return output, fmt.Errorf("%w: %s", err, text)
		}
	}
	return output, err
}
func escapePS(s string) string { return strings.ReplaceAll(s, "'", "''") }
func createShortcut(path, target, args, description, iconPath string) error {
	iconScript := ""
	if strings.TrimSpace(iconPath) != "" {
		iconScript = fmt.Sprintf("$s.IconLocation='%s';", escapePS(iconPath))
	}
	script := fmt.Sprintf("$ErrorActionPreference='Stop';$w=New-Object -ComObject WScript.Shell;$s=$w.CreateShortcut('%s');$s.TargetPath='%s';$s.Arguments='%s';$s.WorkingDirectory='%s';$s.Description='%s';%s$s.Save()",
		escapePS(path), escapePS(target), escapePS(args), escapePS(filepath.Dir(target)), escapePS(description), iconScript)
	_, err := runHiddenTimeout(12*time.Second, "powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	return err
}
func ensureIconFile(dir string) string {
	if strings.TrimSpace(dir) == "" {
		return ""
	}
	_ = os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "NetWatcher.ico")
	if err := os.WriteFile(path, embeddedAppIcon, 0644); err != nil {
		return ""
	}
	return path
}
func runtimeIconPath() string {
	return ensureIconFile(filepath.Join(os.TempDir(), "NetWatcher"))
}
func loadIconFromFile(path string) syscall.Handle {
	if strings.TrimSpace(path) == "" {
		return 0
	}
	h, _, _ := procLoadImageW.Call(0, uintptr(unsafe.Pointer(ptr(path))), IMAGE_ICON, 0, 0, LR_LOADFROMFILE|LR_DEFAULTSIZE)
	return syscall.Handle(h)
}
func defaultInstallDir() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}
	return filepath.Join(base, "Programs", appName)
}
func configDir() string {
	base := os.Getenv("APPDATA")
	if base == "" {
		base = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
	}
	dir := filepath.Join(base, appName)
	_ = os.MkdirAll(dir, 0755)
	return dir
}
func configPath() string      { return filepath.Join(configDir(), "settings.json") }
func defaultLanguage() string { return "en" }
func normalizeTheme(theme string) string {
	if strings.EqualFold(strings.TrimSpace(theme), "dark") {
		return "dark"
	}
	return "light"
}
func defaultConfig() Config {
	return Config{Language: defaultLanguage(), Theme: "light", Interval: 2, TimeoutMS: 1500, HighLatencyMS: 150, ConfirmCycles: 2, StartMinimizedTray: true, AutoCheckUpdates: true, OutageNotifications: true}
}
func loadConfig() Config {
	cfg := defaultConfig()
	data, err := os.ReadFile(configPath())
	if err == nil {
		_ = json.Unmarshal(data, &cfg)
	}
	cfg.Language = normalizeLanguage(cfg.Language)
	cfg.Theme = normalizeTheme(cfg.Theme)
	if cfg.Interval < 0.5 {
		cfg.Interval = 2
	}
	if cfg.TimeoutMS < 200 {
		cfg.TimeoutMS = 1500
	}
	if cfg.HighLatencyMS < 1 {
		cfg.HighLatencyMS = 150
	}
	if cfg.ConfirmCycles < 1 {
		cfg.ConfirmCycles = 2
	}
	return cfg
}
func saveConfig(cfg Config) error {
	cfg.Language = normalizeLanguage(cfg.Language)
	cfg.Theme = normalizeTheme(cfg.Theme)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configDir(), 0755); err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}
func documentsDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "Documents", "NetWatcherLogs")
	_ = os.MkdirAll(dir, 0755)
	return dir
}
func currentExe() string { path, _ := os.Executable(); return path }

func createControl(parent syscall.Handle, class, text string, style uint32, id int) syscall.Handle {
	h, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(ptr(class))), uintptr(unsafe.Pointer(ptr(text))), uintptr(style), 0, 0, 100, 30, uintptr(parent), uintptr(id), 0, 0)
	font, _, _ := procGetStockObject.Call(DEFAULT_GUI_FONT)
	procSendMessageW.Call(h, WM_SETFONT, font, 1)
	return syscall.Handle(h)
}
func createUIFont(height int32, weight int32) syscall.Handle {
	r, _, _ := procCreateFontW.Call(
		uintptr(uint32(height)), 0, 0, 0, uintptr(weight), 0, 0, 0,
		DEFAULT_CHARSET, 0, 0, CLEARTYPE_QUALITY, 0,
		uintptr(unsafe.Pointer(ptr("Segoe UI"))),
	)
	return syscall.Handle(r)
}
func setFont(hwnd syscall.Handle, font syscall.Handle) {
	if hwnd != 0 && font != 0 {
		procSendMessageW.Call(uintptr(hwnd), WM_SETFONT, uintptr(font), 1)
	}
}
func isDarkTheme(theme string) bool { return normalizeTheme(theme) == "dark" }
func applyWindowDarkMode(hwnd syscall.Handle, dark bool) {
	if hwnd == 0 {
		return
	}
	value := int32(0)
	if dark {
		value = 1
	}
	// Attribute 20 is supported by current Windows 10/11 builds; failures are harmless.
	procDwmSetWindowAttribute.Call(uintptr(hwnd), 20, uintptr(unsafe.Pointer(&value)), unsafe.Sizeof(value))
}
func applyControlTheme(hwnd syscall.Handle, dark bool) {
	if hwnd == 0 {
		return
	}
	if dark {
		procSetWindowTheme.Call(uintptr(hwnd), uintptr(unsafe.Pointer(ptr("DarkMode_Explorer"))), 0)
	} else {
		procSetWindowTheme.Call(uintptr(hwnd), uintptr(unsafe.Pointer(ptr("Explorer"))), 0)
	}
	procInvalidateRect.Call(uintptr(hwnd), 0, 1)
}
func fillClientBackground(hwnd syscall.Handle, hdc syscall.Handle, dark bool) {
	var rc RECT
	procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rc)))
	color := rgb(250, 250, 252)
	if dark {
		color = rgb(24, 27, 33)
	}
	brush, _, _ := procCreateSolidBrush.Call(color)
	procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(&rc)), brush)
	procDeleteObject.Call(brush)
}
func themeControlColor(hdc syscall.Handle, dark bool, input bool) uintptr {
	bg := rgb(250, 250, 252)
	fg := rgb(25, 28, 34)
	brush := &lightBackgroundBrush
	if input {
		bg = rgb(255, 255, 255)
		brush = &lightInputBrush
	}
	if dark {
		bg = rgb(35, 39, 47)
		fg = rgb(236, 239, 244)
		brush = &darkBackgroundBrush
		if input {
			brush = &darkInputBrush
		}
	}
	procSetTextColor.Call(uintptr(hdc), fg)
	procSetBkColor.Call(uintptr(hdc), bg)
	if *brush == 0 {
		h, _, _ := procCreateSolidBrush.Call(bg)
		*brush = syscall.Handle(h)
	}
	return uintptr(*brush)
}
func getText(hwnd syscall.Handle) string {
	if hwnd == 0 {
		return ""
	}
	n, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
	buf := make([]uint16, n+1)
	procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}
func setText(hwnd syscall.Handle, text string) {
	if hwnd != 0 {
		procSetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(ptr(text))))
	}
}
func move(hwnd syscall.Handle, x, y, w, h int32) {
	if hwnd != 0 {
		procMoveWindow.Call(uintptr(hwnd), uintptr(x), uintptr(y), uintptr(w), uintptr(h), 1)
	}
}
func show(hwnd syscall.Handle, visible bool) {
	if hwnd != 0 {
		if visible {
			procShowWindow.Call(uintptr(hwnd), SW_SHOW)
		} else {
			procShowWindow.Call(uintptr(hwnd), SW_HIDE)
		}
	}
}
func enable(hwnd syscall.Handle, enabled bool) {
	if hwnd != 0 {
		var value uintptr
		if enabled {
			value = 1
		}
		procEnableWindow.Call(uintptr(hwnd), value)
	}
}
func setCheck(hwnd syscall.Handle, checked bool) {
	value := uintptr(0)
	if checked {
		value = BST_CHECKED
	}
	procSendMessageW.Call(uintptr(hwnd), BM_SETCHECK, value, 0)
}
func isChecked(hwnd syscall.Handle) bool {
	r, _, _ := procSendMessageW.Call(uintptr(hwnd), BM_GETCHECK, 0, 0)
	return r == BST_CHECKED
}
func comboAdd(hwnd syscall.Handle, text string) {
	procSendMessageW.Call(uintptr(hwnd), CB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr(text))))
}
func comboSet(hwnd syscall.Handle, index int) {
	procSendMessageW.Call(uintptr(hwnd), CB_SETCURSEL, uintptr(index), 0)
}
func comboGet(hwnd syscall.Handle) int {
	r, _, _ := procSendMessageW.Call(uintptr(hwnd), CB_GETCURSEL, 0, 0)
	return int(r)
}
func postRefresh(hwnd syscall.Handle) {
	if hwnd != 0 {
		procPostMessageW.Call(uintptr(hwnd), msgRefresh, 0, 0)
	}
}

// ----------------------------- Installer -----------------------------

func folderBrowseProc(hwnd syscall.Handle, msg uint32, lParam, data uintptr) uintptr {
	if msg == BFFM_INITIALIZED && data != 0 {
		procSendMessageW.Call(uintptr(hwnd), BFFM_SETSELECTIONW, 1, data)
	}
	return 0
}

func browseForFolder(owner syscall.Handle, title, initial string) (string, bool) {
	procOleInitialize.Call(0)
	defer procOleUninitialize.Call()

	display := make([]uint16, 260)
	initialPtr := ptr(initial)
	info := BROWSEINFO{
		HwndOwner:      owner,
		PszDisplayName: &display[0],
		LpszTitle:      ptr(title),
		UlFlags:        BIF_RETURNONLYFSDIRS | BIF_EDITBOX | BIF_NEWDIALOGSTYLE,
		Lpfn:           folderBrowseCallback,
		LParam:         uintptr(unsafe.Pointer(initialPtr)),
	}
	pidl, _, _ := procSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(&info)))
	if pidl == 0 {
		return "", false
	}
	defer procCoTaskMemFree.Call(pidl)
	pathBuffer := make([]uint16, 32768)
	ok, _, _ := procSHGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&pathBuffer[0])))
	if ok == 0 {
		return "", false
	}
	path := syscall.UTF16ToString(pathBuffer)
	return strings.TrimSpace(path), path != ""
}

func newInstaller() *Installer {
	return &Installer{
		controls:     map[int]syscall.Handle{},
		pageControls: map[wizard.Page][]syscall.Handle{},
		options:      map[int]bool{instStartMenu: true, instLaunch: true},
		language:     defaultLanguage(),
		installDir:   defaultInstallDir(),
	}
}
func (i *Installer) addPageControl(page wizard.Page, h syscall.Handle) syscall.Handle {
	i.pageControls[page] = append(i.pageControls[page], h)
	return h
}
func (i *Installer) buildControls() {
	i.titleFont = createUIFont(-30, FW_BOLD)
	i.brandFont = createUIFont(-31, FW_BOLD)
	i.taglineFont = createUIFont(-12, FW_SEMIBOLD)
	i.headerFont = createUIFont(-16, FW_SEMIBOLD)
	i.normalFont = createUIFont(-14, FW_NORMAL)
	i.smallFont = createUIFont(-12, FW_NORMAL)

	// The header, page titles, labels and helper copy are painted by the parent
	// window. Only interactive controls are native child windows. This prevents
	// old STATIC control pixels from surviving Back/Next page transitions.
	i.controls[instBack] = createControl(i.hwnd, "BUTTON", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_OWNERDRAW, instBack)
	i.controls[instNext] = createControl(i.hwnd, "BUTTON", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_OWNERDRAW, instNext)
	i.controls[instCancel] = createControl(i.hwnd, "BUTTON", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_OWNERDRAW, instCancel)

	i.controls[instLanguageCombo] = i.addPageControl(wizard.Welcome, createControl(i.hwnd, "COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_VSCROLL|CBS_DROPDOWNLIST, instLanguageCombo))
	comboAdd(i.controls[instLanguageCombo], "Türkçe")
	comboAdd(i.controls[instLanguageCombo], "English")
	if i.language == "en" {
		comboSet(i.controls[instLanguageCombo], 1)
	} else {
		comboSet(i.controls[instLanguageCombo], 0)
	}

	i.controls[instPathEdit] = i.addPageControl(wizard.Options, createControl(i.hwnd, "EDIT", i.installDir, WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL, instPathEdit))
	i.controls[instBrowse] = i.addPageControl(wizard.Options, createControl(i.hwnd, "BUTTON", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_OWNERDRAW, instBrowse))
	i.controls[instDesktop] = i.addPageControl(wizard.Options, createControl(i.hwnd, "BUTTON", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_OWNERDRAW, instDesktop))
	i.controls[instStartMenu] = i.addPageControl(wizard.Options, createControl(i.hwnd, "BUTTON", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_OWNERDRAW, instStartMenu))
	i.controls[instAutoStart] = i.addPageControl(wizard.Options, createControl(i.hwnd, "BUTTON", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_OWNERDRAW, instAutoStart))

	i.controls[instSummary] = i.addPageControl(wizard.Ready, createControl(i.hwnd, "EDIT", "", WS_CHILD|WS_VISIBLE|WS_BORDER|ES_MULTILINE|ES_READONLY|ES_AUTOVSCROLL|WS_VSCROLL, instSummary))
	i.controls[instProgress] = i.addPageControl(wizard.Installing, createControl(i.hwnd, "msctls_progress32", "", WS_CHILD|WS_VISIBLE, instProgress))
	procSendMessageW.Call(uintptr(i.controls[instProgress]), PBM_SETRANGE32, 0, 100)
	i.controls[instDetails] = i.addPageControl(wizard.Installing, createControl(i.hwnd, "EDIT", "", WS_CHILD|WS_VISIBLE|WS_BORDER|ES_MULTILINE|ES_READONLY|ES_AUTOVSCROLL|WS_VSCROLL, instDetails))
	i.controls[instLaunch] = i.addPageControl(wizard.Finished, createControl(i.hwnd, "BUTTON", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_OWNERDRAW, instLaunch))

	for id, h := range i.controls {
		setFont(h, i.normalFont)
		if id != instBack && id != instNext && id != instCancel && id != instBrowse &&
			id != instDesktop && id != instStartMenu && id != instAutoStart && id != instLaunch {
			applyControlTheme(h, false)
		}
	}
	setFont(i.controls[instDetails], i.smallFont)
	i.showPage(wizard.Welcome)
	i.applyLanguage()
}
func (i *Installer) relayout() {
	var rc RECT
	procGetClientRect.Call(uintptr(i.hwnd), uintptr(unsafe.Pointer(&rc)))
	i.layout(rc.Right-rc.Left, rc.Bottom-rc.Top)
}
func (i *Installer) layout(width, height int32) {
	layout := wizard.ComputeLayout(width, height, i.page)
	for id, rect := range layout.Controls {
		move(i.controls[id], rect.X, rect.Y, rect.W, rect.H)
	}
}
func (i *Installer) showPage(page wizard.Page) {
	i.page = page
	visible := wizard.VisibleControls(page)
	for id, h := range i.controls {
		if id == instBack || id == instNext || id == instCancel {
			show(h, true)
			continue
		}
		show(h, visible[id])
	}
	enable(i.controls[instBack], page == wizard.Options || page == wizard.Ready)
	enable(i.controls[instCancel], !i.installing && page != wizard.Finished)
	enable(i.controls[instNext], !i.installing)
	i.applyLanguage()
	i.relayout()
	procRedrawWindow.Call(uintptr(i.hwnd), 0, 0, RDW_INVALIDATE|RDW_ERASE|RDW_ALLCHILDREN|RDW_UPDATENOW)
}
func (i *Installer) applyLanguage() {
	lang := i.language
	setText(i.hwnd, it(lang, "window"))
	setText(i.controls[instBack], it(lang, "back"))
	setText(i.controls[instCancel], it(lang, "cancel"))
	nextText := it(lang, "next")
	if i.page == wizard.Ready {
		nextText = it(lang, "install")
	}
	if i.page == wizard.Finished {
		nextText = it(lang, "finish")
	}
	setText(i.controls[instNext], nextText)
	setText(i.controls[instBrowse], it(lang, "browse"))
	setText(i.controls[instDesktop], it(lang, "desktop"))
	setText(i.controls[instStartMenu], it(lang, "startmenu"))
	setText(i.controls[instAutoStart], it(lang, "autostart"))
	setText(i.controls[instLaunch], it(lang, "launch"))
	if i.page == wizard.Ready {
		i.updateSummary()
	}
	procRedrawWindow.Call(uintptr(i.hwnd), 0, 0, RDW_INVALIDATE|RDW_ERASE|RDW_ALLCHILDREN)
}
func (i *Installer) optionChecked(id int) bool {
	return i.options[id]
}
func (i *Installer) toggleOption(id int) {
	i.options[id] = !i.options[id]
	procInvalidateRect.Call(uintptr(i.controls[id]), 0, 1)
}

func (i *Installer) updateSummary() {
	path := strings.TrimSpace(getText(i.controls[instPathEdit]))
	yesNo := func(v bool) string {
		if v {
			return it(i.language, "yes")
		}
		return it(i.language, "no")
	}
	langName := "Türkçe"
	if i.language == "en" {
		langName = "English"
	}
	summary := fmt.Sprintf("%s\r\n\r\n%s:\r\n%s\r\n\r\n%s: %s\r\n%s: %s\r\n%s: %s\r\n%s: %s",
		it(i.language, "summary_sub"),
		it(i.language, "summary_path"), path,
		it(i.language, "summary_language"), langName,
		it(i.language, "summary_desktop"), yesNo(i.optionChecked(instDesktop)),
		it(i.language, "summary_startmenu"), yesNo(i.optionChecked(instStartMenu)),
		it(i.language, "summary_autostart"), yesNo(i.optionChecked(instAutoStart)))
	setText(i.controls[instSummary], summary)
}
func (i *Installer) browseFolder() {
	current := strings.TrimSpace(getText(i.controls[instPathEdit]))
	initial := current
	if _, err := os.Stat(initial); err != nil {
		initial = filepath.Dir(initial)
	}
	selected, ok := browseForFolder(i.hwnd, it(i.language, "folder_dialog"), initial)
	if !ok {
		return
	}
	if !strings.EqualFold(filepath.Base(selected), appName) {
		selected = filepath.Join(selected, appName)
	}
	setText(i.controls[instPathEdit], selected)
	procSetForegroundWindow.Call(uintptr(i.hwnd))
}
func (i *Installer) appendInstallLog(text string, progress int) {
	i.mu.Lock()
	i.logs = append(i.logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), text))
	i.progress = progress
	i.mu.Unlock()
	procPostMessageW.Call(uintptr(i.hwnd), msgInstallerProgress, 0, 0)
}
func (i *Installer) install() {
	i.mu.Lock()
	i.installing = true
	i.installErr = ""
	i.logs = nil
	i.progress = 0
	i.installDir = strings.TrimSpace(getText(i.controls[instPathEdit]))
	i.mu.Unlock()
	i.showPage(wizard.Installing)
	enable(i.controls[instBack], false)
	enable(i.controls[instNext], false)
	enable(i.controls[instCancel], false)

	go func() {
		finish := func(err error) {
			i.mu.Lock()
			if err != nil {
				i.installErr = err.Error()
			} else {
				i.installed = true
			}
			i.installing = false
			i.mu.Unlock()
			procPostMessageW.Call(uintptr(i.hwnd), msgInstallerDone, 0, 0)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		i.mu.Lock()
		installDir := i.installDir
		i.mu.Unlock()
		var dest, iconDest, self string

		steps := []wizard.Step{
			{
				Name: it(i.language, "detail_prepare"), Progress: 8, Fatal: true,
				Run: func(context.Context) error {
					if installDir == "" || !filepath.IsAbs(installDir) {
						return fmt.Errorf("%s", it(i.language, "invalid_path"))
					}
					if err := os.MkdirAll(installDir, 0755); err != nil {
						return err
					}
					dest = filepath.Join(installDir, appName+".exe")
					iconDest = ensureIconFile(installDir)
					var err error
					self, err = os.Executable()
					if err != nil {
						return err
					}
					i.mu.Lock()
					i.installedExe = dest
					i.mu.Unlock()
					return nil
				},
			},
			{
				Name: it(i.language, "detail_copy"), Progress: 28, Fatal: true,
				Run: func(context.Context) error {
					selfAbs, _ := filepath.Abs(self)
					destAbs, _ := filepath.Abs(dest)
					if strings.EqualFold(filepath.Clean(selfAbs), filepath.Clean(destAbs)) {
						return nil
					}
					if _, err := os.Stat(dest); err == nil {
						_, _ = runHiddenTimeout(4*time.Second, "taskkill", "/IM", appName+".exe", "/F")
						time.Sleep(250 * time.Millisecond)
					}
					tempDest := dest + ".installing"
					_ = os.Remove(tempDest)
					input, err := os.Open(self)
					if err != nil {
						return err
					}
					defer input.Close()
					output, err := os.Create(tempDest)
					if err != nil {
						return err
					}
					_, copyErr := io.Copy(output, input)
					closeErr := output.Close()
					if copyErr != nil {
						_ = os.Remove(tempDest)
						return copyErr
					}
					if closeErr != nil {
						_ = os.Remove(tempDest)
						return closeErr
					}
					_ = os.Remove(dest)
					if err := os.Rename(tempDest, dest); err != nil {
						_ = os.Remove(tempDest)
						return err
					}
					return nil
				},
			},
			{
				Name: it(i.language, "detail_settings"), Progress: 48, Fatal: true,
				Run: func(context.Context) error {
					cfg := loadConfig()
					cfg.Language = i.language
					cfg.AutoStart = i.optionChecked(instAutoStart)
					if err := saveConfig(cfg); err != nil {
						return err
					}
					runKey := `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
					if cfg.AutoStart {
						value := fmt.Sprintf("\"%s\" --app", dest)
						_, err := runHiddenTimeout(6*time.Second, "reg", "add", runKey, "/v", appName, "/t", "REG_SZ", "/d", value, "/f")
						return err
					}
					_, _ = runHiddenTimeout(6*time.Second, "reg", "delete", runKey, "/v", appName, "/f")
					return nil
				},
			},
			{
				Name: it(i.language, "detail_shortcuts"), Progress: 68, Fatal: false,
				Run: func(context.Context) error {
					var failures []string
					addFailure := func(err error) {
						if err != nil {
							failures = append(failures, err.Error())
						}
					}
					appData := os.Getenv("APPDATA")
					if i.optionChecked(instStartMenu) {
						startDir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", appName)
						if err := os.MkdirAll(startDir, 0755); err != nil {
							addFailure(err)
						} else {
							addFailure(createShortcut(filepath.Join(startDir, appName+".lnk"), dest, "--app", "Internet connection and outage monitor", iconDest))
							addFailure(createShortcut(filepath.Join(startDir, "Uninstall "+appName+".lnk"), dest, "--uninstall", "Uninstall NetWatcher", iconDest))
						}
					}
					if i.optionChecked(instDesktop) {
						out, desktopErr := runHiddenTimeout(6*time.Second, "powershell", "-NoProfile", "-NonInteractive", "-Command", "[Environment]::GetFolderPath('Desktop')")
						desktop := strings.TrimSpace(string(out))
						if desktop == "" {
							desktop = filepath.Join(os.Getenv("USERPROFILE"), "Desktop")
						}
						addFailure(desktopErr)
						addFailure(createShortcut(filepath.Join(desktop, appName+".lnk"), dest, "--app", "Internet connection and outage monitor", iconDest))
					}
					if len(failures) > 0 {
						return fmt.Errorf("%s", strings.Join(failures, "; "))
					}
					return nil
				},
			},
			{
				Name: it(i.language, "detail_registry"), Progress: 86, Fatal: true,
				Run: func(context.Context) error {
					uninstallKey := `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\NetWatcher`
					uninstallString := fmt.Sprintf("\"%s\" --uninstall", dest)
					commands := [][]string{
						{"add", uninstallKey, "/v", "DisplayName", "/t", "REG_SZ", "/d", appName, "/f"},
						{"add", uninstallKey, "/v", "DisplayVersion", "/t", "REG_SZ", "/d", appVersion, "/f"},
						{"add", uninstallKey, "/v", "Publisher", "/t", "REG_SZ", "/d", appName, "/f"},
						{"add", uninstallKey, "/v", "InstallLocation", "/t", "REG_SZ", "/d", installDir, "/f"},
						{"add", uninstallKey, "/v", "DisplayIcon", "/t", "REG_SZ", "/d", iconDest, "/f"},
						{"add", uninstallKey, "/v", "UninstallString", "/t", "REG_SZ", "/d", uninstallString, "/f"},
						{"add", uninstallKey, "/v", "NoModify", "/t", "REG_DWORD", "/d", "1", "/f"},
						{"add", uninstallKey, "/v", "NoRepair", "/t", "REG_DWORD", "/d", "1", "/f"},
					}
					for _, args := range commands {
						if _, err := runHiddenTimeout(6*time.Second, "reg", args...); err != nil {
							return err
						}
					}
					return nil
				},
			},
		}

		err := wizard.RunPlan(ctx, steps, func(event wizard.Event) {
			if event.Done {
				i.appendInstallLog(it(i.language, "detail_done"), 100)
				return
			}
			if event.Err != nil {
				i.appendInstallLog(fmt.Sprintf("%s: %v", it(i.language, "detail_warning"), event.Err), event.Progress)
				return
			}
			i.appendInstallLog(event.Name, event.Progress)
		})
		finish(err)
	}()
}
func drawInstallerText(hdc syscall.Handle, x, y int32, text string, color uintptr, font syscall.Handle) {
	old, _, _ := procSelectObject.Call(uintptr(hdc), uintptr(font))
	procSetBkMode.Call(uintptr(hdc), TRANSPARENT)
	procSetTextColor.Call(uintptr(hdc), color)
	u16, _ := syscall.UTF16FromString(text)
	if len(u16) > 1 {
		procTextOutW.Call(uintptr(hdc), uintptr(x), uintptr(y), uintptr(unsafe.Pointer(&u16[0])), uintptr(len(u16)-1))
	}
	procSelectObject.Call(uintptr(hdc), old)
}
func fillInstallerRect(hdc syscall.Handle, rect RECT, color uintptr) {
	brush, _, _ := procCreateSolidBrush.Call(color)
	procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(&rect)), brush)
	procDeleteObject.Call(brush)
}
func drawInstallerLine(hdc syscall.Handle, x1, y1, x2, y2 int32, color uintptr, width int32) {
	pen, _, _ := procCreatePen.Call(PS_SOLID, uintptr(width), color)
	old, _, _ := procSelectObject.Call(uintptr(hdc), pen)
	procMoveToEx.Call(uintptr(hdc), uintptr(x1), uintptr(y1), 0)
	procLineTo.Call(uintptr(hdc), uintptr(x2), uintptr(y2))
	procSelectObject.Call(uintptr(hdc), old)
	procDeleteObject.Call(pen)
}
func drawInstallerTextRect(hdc syscall.Handle, rect RECT, text string, color uintptr, font syscall.Handle, flags uint32) {
	old, _, _ := procSelectObject.Call(uintptr(hdc), uintptr(font))
	procSetBkMode.Call(uintptr(hdc), TRANSPARENT)
	procSetTextColor.Call(uintptr(hdc), color)
	u16, _ := syscall.UTF16FromString(text)
	if len(u16) > 1 {
		r := rect
		procDrawTextW.Call(uintptr(hdc), uintptr(unsafe.Pointer(&u16[0])), ^uintptr(0), uintptr(unsafe.Pointer(&r)), uintptr(flags))
	}
	procSelectObject.Call(uintptr(hdc), old)
}
func drawRoundedBox(hdc syscall.Handle, rect RECT, radius int32, fill, border uintptr, borderWidth int32) {
	brush, _, _ := procCreateSolidBrush.Call(fill)
	pen, _, _ := procCreatePen.Call(PS_SOLID, uintptr(borderWidth), border)
	oldBrush, _, _ := procSelectObject.Call(uintptr(hdc), brush)
	oldPen, _, _ := procSelectObject.Call(uintptr(hdc), pen)
	procRoundRect.Call(uintptr(hdc), uintptr(rect.Left), uintptr(rect.Top), uintptr(rect.Right), uintptr(rect.Bottom), uintptr(radius), uintptr(radius))
	procSelectObject.Call(uintptr(hdc), oldBrush)
	procSelectObject.Call(uintptr(hdc), oldPen)
	procDeleteObject.Call(brush)
	procDeleteObject.Call(pen)
}
func drawCircle(hdc syscall.Handle, left, top, right, bottom int32, fill, border uintptr, borderWidth int32) {
	brush, _, _ := procCreateSolidBrush.Call(fill)
	pen, _, _ := procCreatePen.Call(PS_SOLID, uintptr(borderWidth), border)
	oldBrush, _, _ := procSelectObject.Call(uintptr(hdc), brush)
	oldPen, _, _ := procSelectObject.Call(uintptr(hdc), pen)
	procEllipse.Call(uintptr(hdc), uintptr(left), uintptr(top), uintptr(right), uintptr(bottom))
	procSelectObject.Call(uintptr(hdc), oldBrush)
	procSelectObject.Call(uintptr(hdc), oldPen)
	procDeleteObject.Call(brush)
	procDeleteObject.Call(pen)
}
func drawBrandMark(hdc syscall.Handle, cx, cy int32) {
	// Circular connectivity mark with a heartbeat line, matching the NetWatcher identity.
	navy := rgb(8, 37, 88)
	blue := rgb(22, 119, 255)
	cyan := rgb(0, 194, 255)
	white := rgb(255, 255, 255)
	pen, _, _ := procCreatePen.Call(PS_SOLID, 5, white)
	nullBrush, _, _ := procGetStockObject.Call(NULL_BRUSH)
	oldPen, _, _ := procSelectObject.Call(uintptr(hdc), pen)
	oldBrush, _, _ := procSelectObject.Call(uintptr(hdc), nullBrush)
	procEllipse.Call(uintptr(hdc), uintptr(cx-39), uintptr(cy-39), uintptr(cx+39), uintptr(cy+39))
	procSelectObject.Call(uintptr(hdc), oldBrush)
	procSelectObject.Call(uintptr(hdc), oldPen)
	procDeleteObject.Call(pen)
	// Accent segments.
	drawInstallerLine(hdc, cx+15, cy-35, cx+31, cy-22, blue, 6)
	drawInstallerLine(hdc, cx-33, cy+20, cx-18, cy+34, cyan, 6)
	// Four nodes.
	drawCircle(hdc, cx-7, cy-50, cx+7, cy-36, navy, white, 4)
	drawCircle(hdc, cx+36, cy-7, cx+50, cy+7, navy, white, 4)
	drawCircle(hdc, cx-7, cy+36, cx+7, cy+50, navy, cyan, 4)
	drawCircle(hdc, cx-50, cy-7, cx-36, cy+7, navy, white, 4)
	// Heartbeat waveform.
	pts := [][2]int32{{cx - 31, cy + 2}, {cx - 17, cy + 2}, {cx - 10, cy - 10}, {cx - 1, cy + 23}, {cx + 9, cy - 28}, {cx + 18, cy + 14}, {cx + 26, cy + 2}, {cx + 34, cy + 2}}
	for j := 0; j < len(pts)-1; j++ {
		drawInstallerLine(hdc, pts[j][0], pts[j][1], pts[j+1][0], pts[j+1][1], cyan, 5)
	}
}
func drawNetworkDecoration(hdc syscall.Handle, width int32) {
	line := rgb(46, 139, 255)
	node := rgb(126, 213, 255)
	points := [][2]int32{{width - 330, 34}, {width - 255, 72}, {width - 185, 31}, {width - 118, 88}, {width - 52, 42}, {width - 220, 117}, {width - 80, 128}}
	edges := [][2]int{{0, 1}, {1, 2}, {1, 5}, {2, 3}, {2, 4}, {3, 4}, {3, 5}, {3, 6}, {5, 6}}
	for _, e := range edges {
		a, b := points[e[0]], points[e[1]]
		drawInstallerLine(hdc, a[0], a[1], b[0], b[1], line, 1)
	}
	for _, p := range points {
		drawCircle(hdc, p[0]-4, p[1]-4, p[0]+4, p[1]+4, node, node, 1)
	}
}
func (i *Installer) drawOwnerControl(dis *DRAWITEMSTRUCT) {
	hdc := dis.HDC
	rc := dis.RcItem
	id := int(dis.CtlID)
	text := getText(dis.HwndItem)
	selected := dis.ItemState&ODS_SELECTED != 0
	disabled := dis.ItemState&ODS_DISABLED != 0
	if id == instDesktop || id == instStartMenu || id == instAutoStart || id == instLaunch {
		fillInstallerRect(hdc, rc, rgb(255, 255, 255))
		boxTop := (rc.Bottom - rc.Top - 22) / 2
		box := RECT{8, boxTop, 30, boxTop + 22}
		fill := rgb(255, 255, 255)
		border := rgb(22, 119, 255)
		if i.optionChecked(id) {
			fill = rgb(22, 119, 255)
		}
		drawRoundedBox(hdc, box, 5, fill, border, 1)
		if i.optionChecked(id) {
			drawInstallerLine(hdc, 13, boxTop+11, 18, boxTop+16, rgb(255, 255, 255), 2)
			drawInstallerLine(hdc, 18, boxTop+16, 26, boxTop+7, rgb(255, 255, 255), 2)
		}
		textColor := rgb(17, 34, 63)
		if disabled {
			textColor = rgb(145, 151, 162)
		}
		drawInstallerTextRect(hdc, RECT{44, 0, rc.Right - 4, rc.Bottom}, text, textColor, i.normalFont, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
		return
	}
	primary := id == instNext
	fill := rgb(255, 255, 255)
	border := rgb(190, 201, 216)
	textColor := rgb(16, 37, 70)
	if primary {
		fill = rgb(22, 119, 255)
		border = rgb(22, 119, 255)
		textColor = rgb(255, 255, 255)
	}
	if id == instBrowse {
		border = rgb(22, 119, 255)
		textColor = rgb(14, 91, 190)
	}
	if selected {
		if primary {
			fill = rgb(10, 89, 205)
		} else {
			fill = rgb(234, 242, 255)
		}
	}
	if disabled {
		fill = rgb(237, 240, 244)
		border = rgb(212, 217, 225)
		textColor = rgb(150, 156, 166)
	}
	drawRoundedBox(hdc, RECT{1, 1, rc.Right - 1, rc.Bottom - 1}, 9, fill, border, 1)
	drawInstallerTextRect(hdc, rc, text, textColor, i.headerFont, DT_CENTER|DT_VCENTER|DT_SINGLELINE)
}
func (i *Installer) paint(hdc syscall.Handle, rc RECT) {
	width, height := rc.Right-rc.Left, rc.Bottom-rc.Top
	layout := wizard.ComputeLayout(width, height, i.page)
	toRect := func(r wizard.Rect) RECT { return RECT{r.X, r.Y, r.X + r.W, r.Y + r.H} }

	white := rgb(255, 255, 255)
	blue := rgb(22, 119, 255)
	cyan := rgb(0, 194, 255)
	dark := rgb(11, 19, 43)
	muted := rgb(82, 101, 128)
	border := rgb(222, 229, 238)
	fillInstallerRect(hdc, rc, rgb(248, 250, 253))

	// Branded gradient header.
	headerH := layout.Header.H
	steps := int32(72)
	for x := int32(0); x < steps; x++ {
		t := float64(x) / float64(steps-1)
		r := uint8(5 + 4*t)
		g := uint8(35 + 87*t)
		b := uint8(91 + 140*t)
		left := width * x / steps
		right := width * (x + 1) / steps
		fillInstallerRect(hdc, RECT{left, 0, right, headerH}, rgb(r, g, b))
	}
	drawNetworkDecoration(hdc, width)
	drawBrandMark(hdc, 74, 70)
	drawInstallerText(hdc, 134, 35, "Net", white, i.brandFont)
	drawInstallerText(hdc, 192, 35, "Watcher", cyan, i.brandFont)
	drawInstallerText(hdc, 137, 82, "MONITOR.  CONNECT.  PROTECT.", rgb(220, 239, 255), i.taglineFont)

	fillInstallerRect(hdc, toRect(layout.Footer), rgb(249, 250, 252))
	drawInstallerLine(hdc, 0, layout.Footer.Y, width, layout.Footer.Y, border, 1)

	titleKeys := []string{"welcome_title", "options_title", "summary_title", "installing_title", "finish_title"}
	subtitleKeys := []string{"welcome_sub", "options_sub", "summary_sub", "installing_sub", "finish_sub"}
	pageIndex := int(i.page)
	if pageIndex >= 0 && pageIndex < len(titleKeys) {
		drawInstallerTextRect(hdc, toRect(layout.Title), it(i.language, titleKeys[pageIndex]), dark, i.titleFont, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
		drawInstallerTextRect(hdc, toRect(layout.Subtitle), it(i.language, subtitleKeys[pageIndex]), muted, i.normalFont, DT_LEFT|DT_WORDBREAK)
	}

	drawRoundedBox(hdc, toRect(layout.Card), 18, white, border, 1)

	switch i.page {
	case wizard.Welcome:
		drawInstallerTextRect(hdc, toRect(layout.TextRegions["welcomeBody"]), it(i.language, "welcome_body"), muted, i.normalFont, DT_LEFT|DT_WORDBREAK)
		drawInstallerLine(hdc, 90, 400, width-90, 400, rgb(231, 236, 243), 1)
		drawInstallerTextRect(hdc, toRect(layout.TextRegions["languageLabel"]), it(i.language, "lang_text"), dark, i.headerFont, DT_LEFT|DT_VCENTER|DT_SINGLELINE)
	case wizard.Options:
		folderX, folderY := int32(74), int32(299)
		drawInstallerLine(hdc, folderX, folderY+5, folderX+13, folderY+5, blue, 3)
		drawInstallerLine(hdc, folderX+13, folderY+5, folderX+19, folderY+11, blue, 3)
		drawInstallerLine(hdc, folderX+19, folderY+11, folderX+43, folderY+11, blue, 3)
		drawInstallerLine(hdc, folderX, folderY+5, folderX, folderY+35, blue, 3)
		drawInstallerLine(hdc, folderX, folderY+35, folderX+43, folderY+35, blue, 3)
		drawInstallerLine(hdc, folderX+43, folderY+35, folderX+43, folderY+11, blue, 3)
		drawInstallerTextRect(hdc, toRect(layout.TextRegions["pathLabel"]), it(i.language, "path_label"), dark, i.headerFont, DT_LEFT|DT_VCENTER|DT_SINGLELINE)
		drawInstallerTextRect(hdc, toRect(layout.TextRegions["pathNote"]), it(i.language, "path_note"), muted, i.smallFont, DT_LEFT|DT_WORDBREAK)
		for _, id := range []int{instDesktop, instStartMenu, instAutoStart} {
			r := layout.Controls[id]
			drawInstallerLine(hdc, 86, r.Y-8, width-86, r.Y-8, rgb(231, 236, 243), 1)
		}
	case wizard.Installing:
		drawInstallerTextRect(hdc, toRect(layout.TextRegions["detailsLabel"]), it(i.language, "details"), dark, i.headerFont, DT_LEFT|DT_VCENTER|DT_SINGLELINE)
	case wizard.Finished:
		drawInstallerTextRect(hdc, toRect(layout.TextRegions["finishBody"]), it(i.language, "finish_text"), muted, i.normalFont, DT_LEFT|DT_WORDBREAK)
	}
}
func runInstaller() {
	init := INITCOMMONCONTROLSEX{DwSize: uint32(unsafe.Sizeof(INITCOMMONCONTROLSEX{})), DwICC: ICC_PROGRESS_CLASS}
	procInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&init)))
	globalInstaller = newInstaller()
	// Setup wizards are intentionally fixed-size. This avoids expensive live relayout/redraw
	// loops while dragging the window border and matches standard Windows installers.
	runWindowStyle("NetWatcherInstallerWindow", installerProcCallback, it(globalInstaller.language, "window"), 940, 680, WS_CAPTION|WS_SYSMENU|WS_MINIMIZEBOX|WS_CLIPCHILDREN|WS_VISIBLE, true)
}
func installerProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	i := globalInstaller
	switch msg {
	case WM_CREATE:
		i.hwnd = hwnd
		i.buildControls()
		return 0
	case WM_SIZE:
		i.layout(int32(loword(lParam)), int32(hiword(lParam)))
		return 0
	case WM_ERASEBKGND:
		return 1
	case WM_CTLCOLORSTATIC:
		hdc := syscall.Handle(wParam)
		procSetBkMode.Call(uintptr(hdc), TRANSPARENT)
		procSetTextColor.Call(uintptr(hdc), rgb(17, 34, 63))
		brush, _, _ := procGetStockObject.Call(NULL_BRUSH)
		return brush
	case WM_CTLCOLOREDIT:
		return themeControlColor(syscall.Handle(wParam), false, true)
	case WM_DRAWITEM:
		if lParam != 0 {
			i.drawOwnerControl((*DRAWITEMSTRUCT)(unsafe.Pointer(lParam)))
			return 1
		}
	case WM_PAINT:
		var ps PAINTSTRUCT
		hdc, _, _ := procBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
		var rc RECT
		procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rc)))
		i.paint(syscall.Handle(hdc), rc)
		procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
		return 0
	case WM_COMMAND:
		id, notify := int(loword(wParam)), hiword(wParam)
		if id == instLanguageCombo && notify == CBN_SELCHANGE {
			if comboGet(i.controls[instLanguageCombo]) == 1 {
				i.language = "en"
			} else {
				i.language = "tr"
			}
			i.applyLanguage()
			procInvalidateRect.Call(uintptr(hwnd), 0, 1)
			return 0
		}
		if notify == BN_CLICKED {
			switch id {
			case instDesktop, instStartMenu, instAutoStart, instLaunch:
				i.toggleOption(id)
			case instBack:
				i.showPage(wizard.Back(i.page))
			case instNext:
				path := strings.TrimSpace(getText(i.controls[instPathEdit]))
				validPath := path != "" && filepath.IsAbs(path)
				nextPage, action, err := wizard.Next(i.page, validPath)
				if err != nil {
					messageBox(it(i.language, "window"), it(i.language, "invalid_path"), MB_OK|MB_ICONERROR)
					return 0
				}
				switch action {
				case wizard.StartInstall:
					i.install()
				case wizard.FinishWizard:
					i.mu.Lock()
					installedExe := i.installedExe
					i.mu.Unlock()
					if i.optionChecked(instLaunch) && installedExe != "" {
						_ = exec.Command(installedExe, "--app").Start()
					}
					procDestroyWindow.Call(uintptr(hwnd))
				default:
					i.showPage(nextPage)
				}
			case instCancel:
				if messageBox(it(i.language, "window"), it(i.language, "confirm_cancel"), MB_YESNO|MB_ICONQUESTION) == IDYES {
					procDestroyWindow.Call(uintptr(hwnd))
				}
			case instBrowse:
				i.browseFolder()
			}
		}
		return 0
	case msgInstallerProgress:
		i.mu.Lock()
		logText := strings.Join(append([]string(nil), i.logs...), "\r\n")
		progress := i.progress
		i.mu.Unlock()
		setText(i.controls[instDetails], logText)
		procSendMessageW.Call(uintptr(i.controls[instProgress]), PBM_SETPOS, uintptr(progress), 0)
		return 0
	case msgInstallerDone:
		i.mu.Lock()
		logText := strings.Join(append([]string(nil), i.logs...), "\r\n")
		progress := i.progress
		installErr := i.installErr
		i.mu.Unlock()
		setText(i.controls[instDetails], logText)
		procSendMessageW.Call(uintptr(i.controls[instProgress]), PBM_SETPOS, uintptr(progress), 0)
		if installErr != "" {
			messageBox(it(i.language, "install_failed"), installErr, MB_OK|MB_ICONERROR)
			i.showPage(wizard.Ready)
			enable(i.controls[instNext], true)
			enable(i.controls[instCancel], true)
		} else {
			i.showPage(wizard.Finished)
			enable(i.controls[instNext], true)
			enable(i.controls[instCancel], false)
		}
		return 0
	case WM_CLOSE:
		i.mu.Lock()
		installing := i.installing
		i.mu.Unlock()
		if installing {
			return 0
		}
		if i.page == wizard.Finished {
			procDestroyWindow.Call(uintptr(hwnd))
			return 0
		}
		if messageBox(it(i.language, "window"), it(i.language, "confirm_cancel"), MB_YESNO|MB_ICONQUESTION) == IDYES {
			procDestroyWindow.Call(uintptr(hwnd))
		}
		return 0
	case WM_DESTROY:
		for _, font := range []syscall.Handle{i.titleFont, i.brandFont, i.taglineFont, i.headerFont, i.normalFont, i.smallFont} {
			if font != 0 {
				procDeleteObject.Call(uintptr(font))
			}
		}
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

func installedPathFromRegistry() string {
	out, err := hiddenCommand("reg", "query", `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\NetWatcher`, "/v", "InstallLocation").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "InstallLocation") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					return strings.Join(parts[2:], " ")
				}
			}
		}
	}
	return filepath.Dir(currentExe())
}
func runUninstaller() {
	title := "Uninstall NetWatcher"
	text := "NetWatcher will be removed from your computer.\n\nMeasurements under Documents\\NetWatcherLogs will be preserved.\n\nContinue?"
	done := "NetWatcher was removed. Measurements and settings were preserved."
	if messageBox(title, text, MB_YESNO|MB_ICONQUESTION) != IDYES {
		return
	}
	installDir := installedPathFromRegistry()
	dest := filepath.Join(installDir, appName+".exe")
	appData := os.Getenv("APPDATA")
	_ = os.RemoveAll(filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", appName))
	out, _ := hiddenCommand("powershell", "-NoProfile", "-Command", "[Environment]::GetFolderPath('Desktop')").Output()
	desktop := strings.TrimSpace(string(out))
	if desktop != "" {
		_ = os.Remove(filepath.Join(desktop, appName+".lnk"))
	}
	_ = hiddenCommand("reg", "delete", `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\NetWatcher`, "/f").Run()
	_ = hiddenCommand("reg", "delete", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", appName, "/f").Run()
	messageBox(title, done, MB_OK|MB_ICONINFORMATION)
	command := fmt.Sprintf("ping 127.0.0.1 -n 3 >nul & rmdir /s /q \"%s\"", installDir)
	_ = hiddenCommand("cmd", "/c", command).Start()
	_ = dest
}

// ----------------------------- Application -----------------------------

func newApp() *App {
	cfg := loadConfig()
	a := &App{controls: map[int]syscall.Handle{}, latest: map[string]PingResult{}, history: map[string][]Sample{}, stopCh: make(chan struct{}), logDir: documentsDir(), config: cfg}
	if gateway := getDefaultGateway(); gateway != "" {
		a.targets = append(a.targets, Target{"Modem/Default Gateway", gateway, "local"})
	} else {
		a.addEvent(tr(cfg.Language, "gateway_missing"))
	}
	a.targets = append(a.targets, Target{"Cloudflare", "1.1.1.1", "internet"}, Target{"Google", "8.8.8.8", "internet"})
	for _, host := range cfg.CustomTargets {
		if strings.TrimSpace(host) != "" {
			prefix := "Custom: "
			a.targets = append(a.targets, Target{prefix + host, host, "internet"})
		}
	}
	return a
}
func getDefaultGateway() string {
	script := `(Get-NetRoute -AddressFamily IPv4 -DestinationPrefix '0.0.0.0/0' | Where-Object {$_.NextHop -ne '0.0.0.0'} | Sort-Object RouteMetric | Select-Object -First 1 -ExpandProperty NextHop)`
	out, err := hiddenCommand("powershell", "-NoProfile", "-Command", script).Output()
	if err == nil {
		candidate := strings.TrimSpace(string(out))
		if net.ParseIP(candidate) != nil {
			return candidate
		}
	}
	return ""
}
func (a *App) addEvent(text string) { a.mu.Lock(); defer a.mu.Unlock(); a.addEventLocked(text) }
func (a *App) addEventLocked(text string) {
	line := fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), text)
	a.events = append(a.events, line)
	if len(a.events) > 400 {
		a.events = a.events[len(a.events)-400:]
	}
}
func pingTarget(t Target, timeout int, lang string) PingResult {
	now := time.Now()
	if timeout < 200 {
		timeout = 200
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout+2000)*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ping", "-n", "1", "-w", strconv.Itoa(timeout), t.Host)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: CREATE_NO_WINDOW}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	output := buf.String()
	success := ctx.Err() == nil && err == nil && strings.Contains(strings.ToLower(output), "ttl=")
	latency := 0.0
	if success {
		normalized := strings.ReplaceAll(output, "<1ms", "=0.5ms")
		match := latencyRe.FindStringSubmatch(normalized)
		if len(match) > 1 {
			latency, _ = strconv.ParseFloat(strings.ReplaceAll(match[1], ",", "."), 64)
		}
	}
	msg := tr(lang, "response_ok")
	if !success {
		msg = tr(lang, "ping_error")
	}
	return PingResult{now, t, success, latency, msg}
}
func (a *App) startMonitoring(interval float64, timeout int) {
	a.mu.Lock()
	if a.monitoring {
		a.mu.Unlock()
		return
	}
	a.monitoring = true
	a.stopCh = make(chan struct{})
	a.config.Interval = interval
	a.config.TimeoutMS = timeout
	_ = saveConfig(a.config)
	a.addEventLocked(tr(a.config.Language, "monitor_started_event"))
	a.mu.Unlock()
	go a.monitorLoop(time.Duration(interval*float64(time.Second)), timeout)
}
func (a *App) stopMonitoring(reason string) {
	a.mu.Lock()
	if !a.monitoring {
		a.mu.Unlock()
		return
	}
	a.monitoring = false
	close(a.stopCh)
	if a.active != nil {
		a.active.End = time.Now()
		a.outages = append(a.outages, *a.active)
		a.writeOutageLocked(*a.active)
		a.addEventLocked(fmt.Sprintf("%s [%s] — %s. %s", tr(a.config.Language, "outage_ended"), a.stateLabel(a.active.Category), formatDuration(a.active.End.Sub(a.active.Start), a.config.Language), reason))
		a.active = nil
	}
	a.addEventLocked(tr(a.config.Language, "monitor_stopped_event"))
	a.mu.Unlock()
	postRefresh(a.hwnd)
}
func (a *App) monitorLoop(interval time.Duration, timeout int) {
	if interval < 500*time.Millisecond {
		interval = 500 * time.Millisecond
	}
	for {
		started := time.Now()
		a.mu.RLock()
		targets := append([]Target(nil), a.targets...)
		stopCh := a.stopCh
		lang := a.config.Language
		a.mu.RUnlock()
		results := make([]PingResult, len(targets))
		var wg sync.WaitGroup
		for idx, target := range targets {
			wg.Add(1)
			go func(i int, t Target) { defer wg.Done(); results[i] = pingTarget(t, timeout, lang) }(idx, target)
		}
		wg.Wait()
		a.handleCycle(results)
		postRefresh(a.hwnd)
		wait := interval - time.Since(started)
		if wait < 50*time.Millisecond {
			wait = 50 * time.Millisecond
		}
		select {
		case <-stopCh:
			return
		case <-time.After(wait):
		}
	}
}
func (a *App) handleCycle(results []PingResult) {
	// Keep the state lock limited to in-memory updates. Disk writes can be
	// delayed by antivirus, OneDrive or a busy drive and must never block the UI.
	a.mu.Lock()
	for _, r := range results {
		a.latest[r.Target.Host] = r
		a.results = append(a.results, r)
		h := append(a.history[r.Target.Host], Sample{r.Timestamp, r.Latency, r.Success})
		if len(h) > maxHistory {
			h = h[len(h)-maxHistory:]
		}
		a.history[r.Target.Host] = h
	}
	state, details := a.classify(results)
	a.updateOutageLocked(state, details, time.Now())
	a.mu.Unlock()

	// Logging happens after releasing the shared state lock.
	for _, r := range results {
		a.writeSampleLocked(r)
	}
	a.drainNotifications()
}
func (a *App) classify(results []PingResult) (string, string) {
	var local *PingResult
	internet := []PingResult{}
	for idx := range results {
		if results[idx].Target.Kind == "local" {
			local = &results[idx]
		} else {
			internet = append(internet, results[idx])
		}
	}
	lang := a.config.Language
	if local != nil && !local.Success {
		return "LOCAL_NETWORK", fmt.Sprintf("%s (%s).", tr(lang, "local_error"), local.Target.Host)
	}
	allFailed := len(internet) > 0
	failed := []string{}
	high := []string{}
	for _, r := range internet {
		if r.Success {
			allFailed = false
		} else {
			failed = append(failed, r.Target.Host)
		}
		if r.Success && r.Latency >= a.config.HighLatencyMS {
			high = append(high, fmt.Sprintf("%s=%.0f ms", r.Target.Host, r.Latency))
		}
	}
	if allFailed {
		return "ISP_OUTAGE", tr(lang, "isp_error")
	}
	if len(failed) > 0 {
		return "DEGRADED", tr(lang, "partial_error") + ": " + strings.Join(failed, ", ")
	}
	if len(high) > 0 {
		return "HIGH_LATENCY", tr(lang, "high_latency") + ": " + strings.Join(high, ", ")
	}
	return "ONLINE", tr(lang, "connection_normal")
}
func (a *App) updateOutageLocked(state, details string, now time.Time) {
	lang := a.config.Language
	if state == "ONLINE" {
		a.pendingState = ""
		a.pendingCount = 0
		if a.active != nil {
			a.active.End = now
			finished := *a.active
			a.outages = append(a.outages, finished)
			a.writeOutageLocked(finished)
			a.addEventLocked(fmt.Sprintf("%s [%s] — %s. %s", tr(lang, "outage_ended"), a.stateLabel(finished.Category), formatDuration(finished.End.Sub(finished.Start), lang), tr(lang, "back_normal")))
			a.queueNotificationLocked("Connection restored", "NetWatcher detected that the connection returned to normal.", NIIF_INFO, "")
			a.active = nil
		}
		return
	}
	if a.active != nil && a.active.Category == state {
		return
	}
	if a.active != nil && a.active.Category != state {
		a.active.End = now
		finished := *a.active
		a.outages = append(a.outages, finished)
		a.writeOutageLocked(finished)
		a.addEventLocked(fmt.Sprintf("%s [%s] — %s. %s", tr(lang, "outage_ended"), a.stateLabel(finished.Category), formatDuration(finished.End.Sub(finished.Start), lang), tr(lang, "state_changed")))
		a.active = nil
	}
	if a.pendingState == state {
		a.pendingCount++
	} else {
		a.pendingState = state
		a.pendingCount = 1
	}
	if a.pendingCount >= a.config.ConfirmCycles {
		a.active = &Outage{Start: now, Category: state, Details: details}
		a.addEventLocked(fmt.Sprintf("%s [%s] — %s", tr(lang, "outage_started"), a.stateLabel(state), details))
		a.queueNotificationLocked("Connection problem detected", a.stateLabel(state)+": "+details, NIIF_WARNING, "")
		a.pendingState = ""
		a.pendingCount = 0
	}
}
func (a *App) stateLabel(state string) string {
	keys := map[string]string{"LOCAL_NETWORK": "state_local", "ISP_OUTAGE": "state_isp", "DEGRADED": "state_degraded", "HIGH_LATENCY": "state_high", "ONLINE": "state_online"}
	if key, ok := keys[state]; ok {
		return tr(a.config.Language, key)
	}
	return state
}
func formatDuration(d time.Duration, lang string) string {
	seconds := int(d.Seconds())
	if seconds < 0 {
		seconds = 0
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if lang == "en" {
		if h > 0 {
			return fmt.Sprintf("%d h %d min %d sec", h, m, s)
		}
		if m > 0 {
			return fmt.Sprintf("%d min %d sec", m, s)
		}
		return fmt.Sprintf("%d sec", s)
	}
	if h > 0 {
		return fmt.Sprintf("%d sa %d dk %d sn", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%d dk %d sn", m, s)
	}
	return fmt.Sprintf("%d sn", s)
}
func ensureCSV(path string, header []string) (*os.File, *csv.Writer, error) {
	_, err := os.Stat(path)
	newFile := os.IsNotExist(err)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}
	if newFile {
		_, _ = f.Write([]byte{0xEF, 0xBB, 0xBF})
	}
	w := csv.NewWriter(f)
	w.Comma = ';'
	if newFile {
		_ = w.Write(header)
	}
	return f, w, nil
}
func (a *App) writeSampleLocked(r PingResult) {
	path := filepath.Join(a.logDir, "samples_"+r.Timestamp.Format("2006-01-02")+".csv")
	f, w, err := ensureCSV(path, []string{"timestamp", "name", "host", "target_type", "success", "latency_ms", "message"})
	if err != nil {
		return
	}
	latency := ""
	if r.Success {
		latency = fmt.Sprintf("%.2f", r.Latency)
	}
	_ = w.Write([]string{r.Timestamp.Format(time.RFC3339), r.Target.Name, r.Target.Host, r.Target.Kind, strconv.FormatBool(r.Success), latency, r.Message})
	w.Flush()
	_ = f.Close()
}
func (a *App) writeOutageLocked(o Outage) {
	path := filepath.Join(a.logDir, "outages.csv")
	f, w, err := ensureCSV(path, []string{"start", "end", "duration_seconds", "category", "details"})
	if err != nil {
		return
	}
	_ = w.Write([]string{o.Start.Format(time.RFC3339), o.End.Format(time.RFC3339), strconv.Itoa(int(o.End.Sub(o.Start).Seconds())), o.Category, o.Details})
	w.Flush()
	_ = f.Close()
}
func (a *App) addCustomTarget(host string) bool {
	host = strings.TrimSpace(host)
	if host == "" {
		return false
	}
	a.mu.Lock()
	for _, t := range a.targets {
		if strings.EqualFold(t.Host, host) {
			a.mu.Unlock()
			return false
		}
	}
	a.targets = append(a.targets, Target{"Custom: " + host, host, "internet"})
	a.config.CustomTargets = append(a.config.CustomTargets, host)
	a.addEventLocked(tr(a.config.Language, "target_added") + ": " + host)
	cfg := a.config
	a.mu.Unlock()
	_ = saveConfig(cfg)
	return true
}

func (a *App) removeCustomTarget(host string) bool {
	host = strings.TrimSpace(host)
	if host == "" {
		messageBox(appName, tr(a.config.Language, "select_custom_target"), MB_OK|MB_ICONINFORMATION)
		return false
	}

	a.mu.RLock()
	lang := a.config.Language
	found := findCustomTargetIndex(a.config.CustomTargets, host) >= 0
	a.mu.RUnlock()
	if !found {
		messageBox(appName, tr(lang, "custom_target_not_found"), MB_OK|MB_ICONWARNING)
		return false
	}
	if messageBox(appName, fmt.Sprintf(tr(lang, "remove_confirm"), host), MB_YESNO|MB_ICONQUESTION) != IDYES {
		return false
	}

	a.mu.Lock()
	custom, removed := removeCustomTargetValue(a.config.CustomTargets, host)
	if !removed {
		a.mu.Unlock()
		return false
	}
	a.config.CustomTargets = custom
	targets := a.targets[:0]
	for _, target := range a.targets {
		if strings.HasPrefix(target.Name, "Custom: ") && strings.EqualFold(target.Host, host) {
			continue
		}
		targets = append(targets, target)
	}
	a.targets = targets
	delete(a.latest, host)
	delete(a.history, host)
	a.addEventLocked(tr(a.config.Language, "target_removed") + ": " + host)
	cfg := a.config
	a.mu.Unlock()
	_ = saveConfig(cfg)
	return true
}

func (a *App) refreshCustomTargetCombo(text string) {
	hwnd := a.controls[ctrlCustom]
	if hwnd == 0 {
		return
	}
	a.mu.RLock()
	targets := append([]string(nil), a.config.CustomTargets...)
	a.mu.RUnlock()
	procSendMessageW.Call(uintptr(hwnd), CB_RESETCONTENT, 0, 0)
	for _, target := range targets {
		if strings.TrimSpace(target) != "" {
			comboAdd(hwnd, target)
		}
	}
	setText(hwnd, text)
}
func (a *App) applyAutoStart() {
	path := currentExe()
	value := fmt.Sprintf("\"%s\" %s", path, autoStartArgument(a.config.StartMinimizedTray))
	key := `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	if a.config.AutoStart {
		_ = hiddenCommand("reg", "add", key, "/v", appName, "/t", "REG_SZ", "/d", value, "/f").Run()
	} else {
		_ = hiddenCommand("reg", "delete", key, "/v", appName, "/f").Run()
	}
}

func copyUTF16(dst []uint16, text string) {
	value, err := syscall.UTF16FromString(text)
	if err != nil {
		return
	}
	copy(dst, value)
}

func (a *App) ensureTrayIcon() {
	if a == nil {
		return
	}
	a.trayMu.Lock()
	defer a.trayMu.Unlock()
	if a.hwnd == 0 || a.trayAdded {
		return
	}
	icon := loadIconFromFile(runtimeIconPath())
	if icon == 0 {
		fallback, _, _ := procLoadIconW.Call(0, IDI_APPLICATION)
		icon = syscall.Handle(fallback)
	}
	nid := NOTIFYICONDATA{
		CbSize:           uint32(unsafe.Sizeof(NOTIFYICONDATA{})),
		HWnd:             a.hwnd,
		UID:              trayIconID,
		UFlags:           NIF_MESSAGE | NIF_ICON | NIF_TIP,
		UCallbackMessage: msgTrayIcon,
		HIcon:            icon,
	}
	copyUTF16(nid.SzTip[:], "NetWatcher - Internet connection monitor")
	result, _, _ := procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
	if result != 0 {
		a.trayAdded = true
		a.trayIcon = icon
	}
}

func (a *App) removeTrayIcon() {
	if a == nil {
		return
	}
	a.trayMu.Lock()
	defer a.trayMu.Unlock()
	if !a.trayAdded || a.hwnd == 0 {
		return
	}
	nid := NOTIFYICONDATA{
		CbSize: uint32(unsafe.Sizeof(NOTIFYICONDATA{})),
		HWnd:   a.hwnd,
		UID:    trayIconID,
	}
	procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
	a.trayAdded = false
	a.trayIcon = 0
}

func (a *App) syncTrayIcon() {
	if a == nil {
		return
	}
	if a.startHidden || a.config.CloseToTray || a.config.OutageNotifications {
		a.ensureTrayIcon()
	} else {
		a.removeTrayIcon()
	}
}

func (a *App) showFromTray() {
	if a == nil || a.hwnd == 0 {
		return
	}
	a.startHidden = false
	procShowWindow.Call(uintptr(a.hwnd), SW_SHOWNORMAL)
	procSetForegroundWindow.Call(uintptr(a.hwnd))
	a.syncTrayIcon()
}

func (a *App) showTrayMenu() {
	if a == nil || a.hwnd == 0 {
		return
	}
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)
	procAppendMenuW.Call(menu, MF_STRING, trayOpenID, uintptr(unsafe.Pointer(ptr("Open NetWatcher"))))
	if strings.TrimSpace(githubRepository) != "" {
		procAppendMenuW.Call(menu, MF_STRING, trayCheckUpdateID, uintptr(unsafe.Pointer(ptr("Check for Updates"))))
	}
	procAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(menu, MF_STRING, trayExitID, uintptr(unsafe.Pointer(ptr("Exit NetWatcher"))))
	var point POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&point)))
	procSetForegroundWindow.Call(uintptr(a.hwnd))
	command, _, _ := procTrackPopupMenu.Call(menu, TPM_RIGHTBUTTON|TPM_RETURNCMD, uintptr(point.X), uintptr(point.Y), 0, uintptr(a.hwnd), 0)
	switch int(command) {
	case trayOpenID:
		a.showFromTray()
	case trayCheckUpdateID:
		a.checkForUpdates(true)
	case trayExitID:
		a.exiting = true
		a.removeTrayIcon()
		a.stopMonitoring(tr(a.config.Language, "app_closed"))
		procDestroyWindow.Call(uintptr(a.hwnd))
	}
}

func (a *App) handleTrayMessage(event uint32) {
	switch event {
	case WM_LBUTTONDBLCLK:
		a.showFromTray()
	case WM_RBUTTONUP:
		a.showTrayMenu()
	case NIN_BALLOONUSERCLICK:
		a.mu.Lock()
		url := a.pendingUpdateURL
		a.pendingUpdateURL = ""
		a.mu.Unlock()
		if url != "" {
			_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		}
	}
}

func (a *App) runFirstLaunchQuestions() {
	if a == nil || a.config.FirstRunComplete {
		return
	}
	startAnswer := messageBox(
		"Welcome to NetWatcher",
		"Start NetWatcher automatically when you sign in to Windows?\n\nYou can change this later in Settings.",
		MB_YESNO|MB_ICONQUESTION,
	)
	trayAnswer := messageBox(
		"Welcome to NetWatcher",
		"Keep NetWatcher running in the notification area when you close the window?\n\nChoose Yes to keep monitoring in the background. Double-click the tray icon to reopen NetWatcher, or right-click it to exit.\n\nYou can change this later in Settings.",
		MB_YESNO|MB_ICONQUESTION,
	)
	a.config.Language = "en"
	a.config.AutoStart = startAnswer == IDYES
	a.config.StartMinimizedTray = true
	a.config.CloseToTray = trayAnswer == IDYES
	a.config.AutoCheckUpdates = true
	a.config.OutageNotifications = true
	a.config.FirstRunComplete = true
	if err := saveConfig(a.config); err != nil {
		a.config.FirstRunComplete = false
		messageBox(appName, "Your startup preferences could not be saved:\n\n"+err.Error(), MB_OK|MB_ICONERROR)
		return
	}
	a.applyAutoStart()
}
func (a *App) report() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.results) == 0 {
		return ""
	}
	lang := a.config.Language
	byHost := map[string][]PingResult{}
	for _, r := range a.results {
		byHost[r.Target.Host] = append(byHost[r.Target.Host], r)
	}
	hosts := make([]string, 0, len(byHost))
	for h := range byHost {
		hosts = append(hosts, h)
	}
	sort.Strings(hosts)
	var rows strings.Builder
	for _, hostName := range hosts {
		items := byHost[hostName]
		failures := 0
		latencies := []float64{}
		for _, r := range items {
			if r.Success {
				latencies = append(latencies, r.Latency)
			} else {
				failures++
			}
		}
		avg, p95 := 0.0, 0.0
		if len(latencies) > 0 {
			sort.Float64s(latencies)
			for _, v := range latencies {
				avg += v
			}
			avg /= float64(len(latencies))
			p95 = latencies[int(float64(len(latencies)-1)*0.95)]
		}
		loss := float64(failures) / float64(len(items)) * 100
		fmt.Fprintf(&rows, "<tr><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td>%.2f%%</td><td>%.2f ms</td><td>%.2f ms</td></tr>", html.EscapeString(items[0].Target.Name), html.EscapeString(hostName), len(items), failures, loss, avg, p95)
	}
	var outageRows strings.Builder
	total := time.Duration(0)
	for _, o := range a.outages {
		total += o.End.Sub(o.Start)
		fmt.Fprintf(&outageRows, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>", o.Start.Format("2006-01-02 15:04:05"), o.End.Format("2006-01-02 15:04:05"), formatDuration(o.End.Sub(o.Start), lang), html.EscapeString(a.stateLabel(o.Category)), html.EscapeString(o.Details))
	}
	empty := outageRows.String()
	if empty == "" {
		empty = `<tr><td colspan="5">` + tr(lang, "no_outage") + `</td></tr>`
	}
	first, last := a.results[0].Timestamp, a.results[len(a.results)-1].Timestamp
	htmlLang := "tr"
	if lang == "en" {
		htmlLang = "en"
	}
	report := fmt.Sprintf(`<!doctype html><html lang="%s"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>%s</title><style>%s</style></head><body><div class="wrap"><section class="hero"><div class="hero-row"><div><h1>%s</h1><p>%s: %s</p></div><button class="print-button" onclick="window.print()">%s</button></div></section><section class="summary-grid"><div class="metric"><span>%s</span><strong>%s — %s</strong></div><div class="metric"><span>%s</span><strong>%d</strong><small>%s: %d</small></div><div class="metric"><span>%s</span><strong>%s</strong></div></section><section class="card"><h2>%s</h2><div class="table-wrap"><table><thead><tr><th>%s</th><th>%s</th><th>%s</th><th>%s</th><th>%s</th><th>%s</th><th>P95</th></tr></thead><tbody>%s</tbody></table></div></section><section class="card"><h2>%s</h2><div class="table-wrap"><table><thead><tr><th>%s</th><th>%s</th><th>%s</th><th>%s</th><th>%s</th></tr></thead><tbody>%s</tbody></table></div></section><div class="note">%s</div></div></body></html>`, htmlLang, tr(lang, "report_title"), reportCSS, tr(lang, "report_title"), tr(lang, "created"), time.Now().Format("2006-01-02 15:04:05"), tr(lang, "print_pdf"), tr(lang, "measurement_range"), first.Format("2006-01-02 15:04:05"), last.Format("2006-01-02 15:04:05"), tr(lang, "total_samples"), len(a.results), tr(lang, "completed_outages"), len(a.outages), tr(lang, "total_outage"), formatDuration(total, lang), tr(lang, "target_summary"), tr(lang, "target"), tr(lang, "address"), tr(lang, "sample"), tr(lang, "failed"), tr(lang, "packet_loss"), tr(lang, "average"), rows.String(), tr(lang, "outage_events"), tr(lang, "start_time"), tr(lang, "end_time"), tr(lang, "duration"), tr(lang, "class"), tr(lang, "description"), empty, tr(lang, "report_note"))
	path := filepath.Join(a.logDir, "netwatcher_report_"+time.Now().Format("20060102_150405")+".html")
	if os.WriteFile(path, []byte(report), 0644) != nil {
		return ""
	}
	return path
}
func (a *App) buildControls() {
	lang := a.config.Language
	a.controls[ctrlStart] = createControl(a.hwnd, "BUTTON", tr(lang, "start"), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_DEFPUSHBUTTON, ctrlStart)
	a.controls[ctrlStop] = createControl(a.hwnd, "BUTTON", tr(lang, "stop"), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, ctrlStop)
	a.controls[ctrlReport] = createControl(a.hwnd, "BUTTON", tr(lang, "report"), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, ctrlReport)
	a.controls[ctrlLogs] = createControl(a.hwnd, "BUTTON", tr(lang, "logs"), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, ctrlLogs)
	a.controls[ctrlSettings] = createControl(a.hwnd, "BUTTON", tr(lang, "settings"), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, ctrlSettings)
	a.controls[ctrlStats] = createControl(a.hwnd, "BUTTON", "Statistics", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, ctrlStats)
	a.controls[ctrlExport] = createControl(a.hwnd, "BUTTON", "Export ZIP", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, ctrlExport)
	a.controls[staticCustom] = createControl(a.hwnd, "STATIC", tr(lang, "custom"), WS_CHILD|WS_VISIBLE|SS_LEFT, staticCustom)
	a.controls[ctrlCustom] = createControl(a.hwnd, "COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|WS_VSCROLL|CBS_DROPDOWN|CBS_AUTOHSCROLL, ctrlCustom)
	for _, target := range a.config.CustomTargets {
		if strings.TrimSpace(target) != "" {
			comboAdd(a.controls[ctrlCustom], target)
		}
	}
	a.controls[ctrlAdd] = createControl(a.hwnd, "BUTTON", tr(lang, "add"), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, ctrlAdd)
	a.controls[ctrlRemove] = createControl(a.hwnd, "BUTTON", tr(lang, "remove"), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, ctrlRemove)
	a.controls[ctrlStatusText] = createControl(a.hwnd, "EDIT", "", WS_CHILD|WS_VISIBLE|WS_BORDER|ES_MULTILINE|ES_READONLY|WS_VSCROLL, ctrlStatusText)
	a.controls[ctrlEventText] = createControl(a.hwnd, "EDIT", "", WS_CHILD|WS_VISIBLE|WS_BORDER|ES_MULTILINE|ES_READONLY|ES_AUTOVSCROLL|WS_VSCROLL, ctrlEventText)
	a.controls[ctrlSummary] = createControl(a.hwnd, "STATIC", tr(lang, "not_started"), WS_CHILD|WS_VISIBLE|SS_LEFT, ctrlSummary)
}
func (a *App) applyLanguage() {
	lang := a.config.Language
	setText(a.hwnd, appName+" "+appVersion)
	setText(a.controls[ctrlStart], tr(lang, "start"))
	setText(a.controls[ctrlStop], tr(lang, "stop"))
	setText(a.controls[ctrlReport], tr(lang, "report"))
	setText(a.controls[ctrlLogs], tr(lang, "logs"))
	setText(a.controls[ctrlSettings], tr(lang, "settings"))
	setText(a.controls[ctrlStats], "Statistics")
	setText(a.controls[ctrlExport], "Export ZIP")
	setText(a.controls[staticCustom], tr(lang, "custom"))
	setText(a.controls[ctrlAdd], tr(lang, "add"))
	setText(a.controls[ctrlRemove], tr(lang, "remove"))
	a.refreshUI()
}
func (a *App) applyTheme() {
	a.mu.RLock()
	dark := isDarkTheme(a.config.Theme)
	a.mu.RUnlock()
	applyWindowDarkMode(a.hwnd, dark)
	for _, h := range a.controls {
		applyControlTheme(h, dark)
	}
	procInvalidateRect.Call(uintptr(a.hwnd), 0, 1)
}
func (a *App) layout(width, height int32) {
	move(a.controls[ctrlStart], 10, 10, 145, 30)
	move(a.controls[ctrlStop], 162, 10, 90, 30)
	move(a.controls[ctrlSettings], width-322, 10, 104, 30)
	move(a.controls[ctrlLogs], width-210, 10, 95, 30)
	move(a.controls[ctrlReport], width-108, 10, 98, 30)
	move(a.controls[ctrlStats], width-210, 47, 95, 28)
	move(a.controls[ctrlExport], width-108, 47, 98, 28)
	move(a.controls[staticCustom], 10, 52, 85, 20)
	move(a.controls[ctrlCustom], 98, 48, 205, 180)
	move(a.controls[ctrlAdd], 310, 47, 95, 28)
	move(a.controls[ctrlRemove], 411, 47, 105, 28)
	statusW := int32(410)
	if width < 900 {
		statusW = width/2 - 15
	}
	move(a.controls[ctrlStatusText], 10, 82, statusW, 250)
	eventY := int32(345)
	move(a.controls[ctrlEventText], 10, eventY, width-20, height-eventY-45)
	move(a.controls[ctrlSummary], 10, height-30, width-20, 22)
}
func (a *App) refreshUI() {
	// Take a snapshot first, then release the lock before calling Win32 APIs.
	// SendMessage/SetWindowText are synchronous and must not run while holding
	// the monitor state lock.
	a.mu.RLock()
	lang := a.config.Language
	targets := append([]Target(nil), a.targets...)
	latest := make(map[string]PingResult, len(a.latest))
	for k, v := range a.latest {
		latest[k] = v
	}
	events := append([]string(nil), a.events...)
	outages := append([]Outage(nil), a.outages...)
	var active *Outage
	if a.active != nil {
		copyActive := *a.active
		active = &copyActive
	}
	monitoring := a.monitoring
	resultCount := len(a.results)
	customTargetCount := len(a.config.CustomTargets)
	a.mu.RUnlock()

	// Keep command buttons synchronized with the real monitor state. This is
	// applied from the UI thread by refreshUI for manual start/stop, automatic
	// monitoring on launch, and tray-start scenarios.
	startEnabled, stopEnabled := monitorButtonState(monitoring)
	enable(a.controls[ctrlStart], startEnabled)
	enable(a.controls[ctrlStop], stopEnabled)
	enable(a.controls[ctrlRemove], customTargetCount > 0)

	var b strings.Builder
	fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%s\t%s\r\n", tr(lang, "target"), tr(lang, "address"), tr(lang, "type"), tr(lang, "status"), tr(lang, "latency"), tr(lang, "last"))
	for _, t := range targets {
		r, ok := latest[t.Host]
		status, latency, last := tr(lang, "waiting"), "-", "-"
		if ok {
			if r.Success {
				status = tr(lang, "online")
				latency = fmt.Sprintf("%.1f ms", r.Latency)
			} else {
				status = tr(lang, "failed")
			}
			last = r.Timestamp.Format("15:04:05")
		}
		kind := tr(lang, "internet")
		if t.Kind == "local" {
			kind = tr(lang, "local")
		}
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%s\t%s\r\n", t.Name, t.Host, kind, status, latency, last)
	}
	setText(a.controls[ctrlStatusText], b.String())
	setText(a.controls[ctrlEventText], strings.Join(events, "\r\n"))
	total := time.Duration(0)
	for _, o := range outages {
		total += o.End.Sub(o.Start)
	}
	activeText := ""
	if active != nil {
		activeText = " | " + tr(lang, "active_outage")
	}
	state := tr(lang, "monitor_stopped")
	if monitoring {
		state = tr(lang, "monitor_running")
	}
	summary := fmt.Sprintf("%s | %s: %d | %s: %d | %s: %s%s", state, tr(lang, "samples"), resultCount, tr(lang, "outages"), len(outages), tr(lang, "total"), formatDuration(total, lang), activeText)
	setText(a.controls[ctrlSummary], summary)
	procInvalidateRect.Call(uintptr(a.hwnd), 0, 0)
}
func drawText(hdc syscall.Handle, x, y int32, text string, color uintptr) {
	procSetTextColor.Call(uintptr(hdc), color)
	procSetBkMode.Call(uintptr(hdc), TRANSPARENT)
	u := syscall.StringToUTF16(text)
	if len(u) > 0 {
		procTextOutW.Call(uintptr(hdc), uintptr(x), uintptr(y), uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)-1))
	}
}

func drawTextClipped(hdc syscall.Handle, rect RECT, text string, color uintptr, flags uint32) {
	procSetTextColor.Call(uintptr(hdc), color)
	procSetBkMode.Call(uintptr(hdc), TRANSPARENT)
	u, _ := syscall.UTF16FromString(text)
	if len(u) > 1 {
		r := rect
		procDrawTextW.Call(uintptr(hdc), uintptr(unsafe.Pointer(&u[0])), ^uintptr(0), uintptr(unsafe.Pointer(&r)), uintptr(flags))
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func (a *App) drawGraph(hdc syscall.Handle, client RECT) {
	// Copy graph data before painting so GDI work never holds the shared lock.
	a.mu.RLock()
	darkTheme := isDarkTheme(a.config.Theme)
	lang := a.config.Language
	targets := append([]Target(nil), a.targets...)
	history := make(map[string][]Sample, len(a.history))
	for host, samples := range a.history {
		history[host] = append([]Sample(nil), samples...)
	}
	a.mu.RUnlock()

	graph := RECT{430, 82, client.Right - 10, 332}
	if graph.Right <= graph.Left+100 {
		return
	}
	graphBG := rgb(16, 21, 28)
	gridColor := rgb(43, 52, 64)
	axisColor := rgb(170, 180, 192)
	legendColor := rgb(216, 222, 233)
	if !darkTheme {
		graphBG = rgb(247, 249, 252)
		gridColor = rgb(213, 219, 227)
		axisColor = rgb(82, 91, 104)
		legendColor = rgb(40, 46, 55)
	}
	brush, _, _ := procCreateSolidBrush.Call(graphBG)
	procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(&graph)), brush)
	procDeleteObject.Call(brush)

	legendTargets := make([]Target, 0, len(targets))
	for _, target := range targets {
		if len(history[target.Host]) > 0 {
			legendTargets = append(legendTargets, target)
		}
	}

	marginL, marginR, marginT := int32(48), int32(12), int32(16)
	availableW := graph.Right - graph.Left - marginL - marginR
	legendCols := 2
	if availableW < 430 {
		legendCols = 1
	} else if availableW >= 820 {
		legendCols = 3
	}
	legendRows := 0
	if len(legendTargets) > 0 {
		legendRows = (len(legendTargets) + legendCols - 1) / legendCols
	}
	marginB := int32(30)
	if needed := int32(12 + legendRows*18); needed > marginB {
		marginB = needed
	}

	plotW := graph.Right - graph.Left - marginL - marginR
	plotH := graph.Bottom - graph.Top - marginT - marginB
	if plotW <= 10 || plotH <= 10 {
		return
	}
	maxLatency := 100.0
	for _, samples := range history {
		for _, sample := range samples {
			if sample.Success && sample.Latency > maxLatency {
				maxLatency = sample.Latency
			}
		}
	}
	maxLatency *= 1.25
	if maxLatency > 1000 {
		maxLatency = 1000
	}
	gridPen, _, _ := procCreatePen.Call(PS_SOLID, 1, gridColor)
	old, _, _ := procSelectObject.Call(uintptr(hdc), gridPen)
	for n := 0; n < 5; n++ {
		y := graph.Top + marginT + int32(float64(plotH)*float64(n)/4)
		procMoveToEx.Call(uintptr(hdc), uintptr(graph.Left+marginL), uintptr(y), 0)
		procLineTo.Call(uintptr(hdc), uintptr(graph.Right-marginR), uintptr(y))
		drawText(hdc, graph.Left+4, y-7, fmt.Sprintf("%.0f", maxLatency*(1-float64(n)/4)), axisColor)
	}
	procSelectObject.Call(uintptr(hdc), old)
	procDeleteObject.Call(gridPen)

	palette := []uintptr{rgb(95, 211, 141), rgb(99, 164, 255), rgb(240, 179, 90), rgb(224, 108, 117), rgb(198, 120, 221), rgb(86, 182, 194)}
	for index, t := range targets {
		samples := history[t.Host]
		if len(samples) == 0 {
			continue
		}
		pen, _, _ := procCreatePen.Call(PS_SOLID, 2, palette[index%len(palette)])
		oldPen, _, _ := procSelectObject.Call(uintptr(hdc), pen)
		started := false
		denominator := maxHistory - 1
		if len(samples) > 1 {
			denominator = len(samples) - 1
		}
		if denominator < 1 {
			denominator = 1
		}
		for n, sample := range samples {
			x := graph.Left + marginL + int32(float64(plotW)*float64(n)/float64(denominator))
			if sample.Success {
				y := graph.Top + marginT + int32(float64(plotH)*(1-minFloat(sample.Latency, maxLatency)/maxLatency))
				if !started {
					procMoveToEx.Call(uintptr(hdc), uintptr(x), uintptr(y), 0)
					started = true
				} else {
					procLineTo.Call(uintptr(hdc), uintptr(x), uintptr(y))
				}
			} else {
				started = false
				failPen, _, _ := procCreatePen.Call(PS_SOLID, 1, rgb(255, 92, 92))
				prev, _, _ := procSelectObject.Call(uintptr(hdc), failPen)
				procMoveToEx.Call(uintptr(hdc), uintptr(x), uintptr(graph.Top+marginT), 0)
				procLineTo.Call(uintptr(hdc), uintptr(x), uintptr(graph.Top+marginT+plotH))
				procSelectObject.Call(uintptr(hdc), prev)
				procDeleteObject.Call(failPen)
			}
		}
		procSelectObject.Call(uintptr(hdc), oldPen)
		procDeleteObject.Call(pen)
	}

	if len(legendTargets) > 0 {
		cellW := plotW / int32(legendCols)
		legendTop := graph.Bottom - int32(legendRows*18) - 4
		for index, target := range legendTargets {
			col := index % legendCols
			row := index / legendCols
			left := graph.Left + marginL + int32(col)*cellW
			top := legendTop + int32(row)*18
			right := left + cellW - 8
			colorIndex := 0
			for originalIndex, originalTarget := range targets {
				if originalTarget.Host == target.Host {
					colorIndex = originalIndex
					break
				}
			}
			legendBrush, _, _ := procCreateSolidBrush.Call(palette[colorIndex%len(palette)])
			oldBrush, _, _ := procSelectObject.Call(uintptr(hdc), legendBrush)
			procRectangle.Call(uintptr(hdc), uintptr(left), uintptr(top+2), uintptr(left+11), uintptr(top+13))
			procSelectObject.Call(uintptr(hdc), oldBrush)
			procDeleteObject.Call(legendBrush)
			textRect := RECT{left + 16, top - 1, right, top + 17}
			drawTextClipped(hdc, textRect, target.Name+" ("+target.Host+")", legendColor, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
		}
	}

	failRect := RECT{graph.Right - 235, graph.Top + 2, graph.Right - 12, graph.Top + 22}
	drawTextClipped(hdc, failRect, tr(lang, "graph_fail"), rgb(255, 138, 138), DT_RIGHT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
}

func openSettings(a *App) {
	if globalSettings != nil && globalSettings.hwnd != 0 {
		procShowWindow.Call(uintptr(globalSettings.hwnd), SW_SHOW)
		procSetForegroundWindow.Call(uintptr(globalSettings.hwnd))
		return
	}

	s := &SettingsWindow{controls: map[int]syscall.Handle{}, parent: a, originalLang: a.config.Language, originalTheme: a.config.Theme}
	globalSettings = s

	instance, _, _ := procGetModuleHandleW.Call(0)
	className := ptr("NetWatcherSettingsWindow")
	icon := loadIconFromFile(runtimeIconPath())
	if icon == 0 {
		fallback, _, _ := procLoadIconW.Call(0, IDI_APPLICATION)
		icon = syscall.Handle(fallback)
	}
	cursor, _, _ := procLoadCursorW.Call(0, IDC_ARROW)
	wc := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		LpfnWndProc:   settingsProcCallback,
		HInstance:     syscall.Handle(instance),
		HIcon:         icon,
		HCursor:       syscall.Handle(cursor),
		HbrBackground: syscall.Handle(COLOR_WINDOW + 1),
		LpszClassName: className,
		HIconSm:       icon,
	}
	atom, _, registerErr := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		if errno, ok := registerErr.(syscall.Errno); !ok || errno != syscall.Errno(1410) {
			globalSettings = nil
			messageBoxOwned(a.hwnd, appName, "Settings window class error: "+registerErr.Error(), MB_OK|MB_ICONERROR)
			return
		}
	}

	const dialogWidth = 640
	const dialogHeight = 610
	style := uint32(WS_CAPTION | WS_SYSMENU | WS_CLIPCHILDREN)
	exStyle := uint32(WS_EX_DLGMODALFRAME | WS_EX_CONTROLPARENT)
	hwnd, _, createErr := procCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(ptr("NetWatcher Settings"))),
		uintptr(style),
		0, 0, dialogWidth, dialogHeight,
		uintptr(a.hwnd), // owner: keeps the dialog above NetWatcher and out of the taskbar
		0, instance, 0,
	)
	if hwnd == 0 {
		globalSettings = nil
		messageBoxOwned(a.hwnd, appName, "Settings window error: "+createErr.Error(), MB_OK|MB_ICONERROR)
		return
	}

	procSendMessageW.Call(hwnd, WM_SETICON, ICON_BIG, uintptr(icon))
	procSendMessageW.Call(hwnd, WM_SETICON, ICON_SMALL, uintptr(icon))

	var parentRect RECT
	if ok, _, _ := procGetWindowRect.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&parentRect))); ok != 0 {
		x := parentRect.Left + (parentRect.Right-parentRect.Left-dialogWidth)/2
		y := parentRect.Top + (parentRect.Bottom-parentRect.Top-dialogHeight)/2
		procMoveWindow.Call(hwnd, uintptr(x), uintptr(y), dialogWidth, dialogHeight, 0)
	}

	enable(a.hwnd, false)
	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)
	procSetForegroundWindow.Call(hwnd)
}
func (s *SettingsWindow) buildControls() {
	a := s.parent
	lang := "en"
	s.controls[setLabelTheme] = createControl(s.hwnd, "STATIC", tr(lang, "theme"), WS_CHILD|WS_VISIBLE|SS_LEFT, setLabelTheme)
	s.controls[setTheme] = createControl(s.hwnd, "COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_VSCROLL|CBS_DROPDOWNLIST, setTheme)
	comboAdd(s.controls[setTheme], tr(lang, "theme_light"))
	comboAdd(s.controls[setTheme], tr(lang, "theme_dark"))
	if isDarkTheme(a.config.Theme) {
		comboSet(s.controls[setTheme], 1)
	} else {
		comboSet(s.controls[setTheme], 0)
	}
	s.controls[setLabelInterval] = createControl(s.hwnd, "STATIC", tr(lang, "interval"), WS_CHILD|WS_VISIBLE|SS_LEFT, setLabelInterval)
	s.controls[setInterval] = createControl(s.hwnd, "EDIT", strconv.FormatFloat(a.config.Interval, 'f', 1, 64), WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|ES_LEFT, setInterval)
	s.controls[setLabelTimeout] = createControl(s.hwnd, "STATIC", tr(lang, "timeout"), WS_CHILD|WS_VISIBLE|SS_LEFT, setLabelTimeout)
	s.controls[setTimeout] = createControl(s.hwnd, "EDIT", strconv.Itoa(a.config.TimeoutMS), WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|ES_LEFT, setTimeout)
	s.controls[setLabelLatency] = createControl(s.hwnd, "STATIC", tr(lang, "high_latency_label"), WS_CHILD|WS_VISIBLE|SS_LEFT, setLabelLatency)
	s.controls[setLatency] = createControl(s.hwnd, "EDIT", strconv.FormatFloat(a.config.HighLatencyMS, 'f', 0, 64), WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|ES_LEFT, setLatency)
	s.controls[setLabelConfirm] = createControl(s.hwnd, "STATIC", tr(lang, "confirm_label"), WS_CHILD|WS_VISIBLE|SS_LEFT, setLabelConfirm)
	s.controls[setConfirm] = createControl(s.hwnd, "EDIT", strconv.Itoa(a.config.ConfirmCycles), WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|ES_LEFT, setConfirm)
	s.controls[setAutoStart] = createControl(s.hwnd, "BUTTON", tr(lang, "auto_start"), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, setAutoStart)
	setCheck(s.controls[setAutoStart], a.config.AutoStart)
	s.controls[setStartMinimized] = createControl(s.hwnd, "BUTTON", tr(lang, "start_minimized_tray"), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, setStartMinimized)
	setCheck(s.controls[setStartMinimized], a.config.StartMinimizedTray)
	enable(s.controls[setStartMinimized], a.config.AutoStart)
	s.controls[setAutoMonitor] = createControl(s.hwnd, "BUTTON", tr(lang, "auto_monitor"), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, setAutoMonitor)
	setCheck(s.controls[setAutoMonitor], a.config.AutoMonitor)
	s.controls[setCloseToTray] = createControl(s.hwnd, "BUTTON", tr(lang, "close_to_tray"), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, setCloseToTray)
	setCheck(s.controls[setCloseToTray], a.config.CloseToTray)
	s.controls[setAutoUpdate] = createControl(s.hwnd, "BUTTON", "Automatically check GitHub for updates", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, setAutoUpdate)
	setCheck(s.controls[setAutoUpdate], a.config.AutoCheckUpdates)
	enable(s.controls[setAutoUpdate], strings.TrimSpace(githubRepository) != "")
	s.controls[setOutageNotify] = createControl(s.hwnd, "BUTTON", "Show Windows notifications for outages and recovery", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, setOutageNotify)
	setCheck(s.controls[setOutageNotify], a.config.OutageNotifications)
	s.controls[setInfo] = createControl(s.hwnd, "STATIC", "NetWatcher uses English for the interface, reports and log messages.", WS_CHILD|WS_VISIBLE|SS_LEFT, setInfo)
	s.controls[setSave] = createControl(s.hwnd, "BUTTON", tr(lang, "save"), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_DEFPUSHBUTTON, setSave)
	s.controls[setCancel] = createControl(s.hwnd, "BUTTON", tr(lang, "cancel"), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, setCancel)
}
func (s *SettingsWindow) layout(width, height int32) {
	labelX := int32(28)
	editX := int32(310)
	move(s.controls[setLabelTheme], labelX, 28, 260, 24)
	move(s.controls[setTheme], editX, 24, 200, 160)
	move(s.controls[setLabelInterval], labelX, 72, 260, 24)
	move(s.controls[setInterval], editX, 68, 110, 26)
	move(s.controls[setLabelTimeout], labelX, 112, 260, 24)
	move(s.controls[setTimeout], editX, 108, 110, 26)
	move(s.controls[setLabelLatency], labelX, 152, 260, 24)
	move(s.controls[setLatency], editX, 148, 110, 26)
	move(s.controls[setLabelConfirm], labelX, 192, 260, 24)
	move(s.controls[setConfirm], editX, 188, 110, 26)
	move(s.controls[setAutoStart], labelX, 238, width-56, 28)
	move(s.controls[setStartMinimized], labelX+24, 274, width-80, 28)
	move(s.controls[setAutoMonitor], labelX, 310, width-56, 28)
	move(s.controls[setCloseToTray], labelX, 346, width-56, 34)
	move(s.controls[setAutoUpdate], labelX, 382, width-56, 28)
	move(s.controls[setOutageNotify], labelX, 416, width-56, 28)
	move(s.controls[setInfo], labelX, 454, width-56, 38)
	move(s.controls[setSave], width-194, height-48, 80, 30)
	move(s.controls[setCancel], width-106, height-48, 80, 30)
}
func (s *SettingsWindow) applyLanguage() {
	lang := "en"
	setText(s.hwnd, tr(lang, "settings_title"))
	setText(s.controls[setLabelTheme], tr(lang, "theme"))
	currentTheme := comboGet(s.controls[setTheme])
	procSendMessageW.Call(uintptr(s.controls[setTheme]), 0x014B, 0, 0)
	comboAdd(s.controls[setTheme], tr(lang, "theme_light"))
	comboAdd(s.controls[setTheme], tr(lang, "theme_dark"))
	comboSet(s.controls[setTheme], currentTheme)
	setText(s.controls[setLabelInterval], tr(lang, "interval"))
	setText(s.controls[setLabelTimeout], tr(lang, "timeout"))
	setText(s.controls[setLabelLatency], tr(lang, "high_latency_label"))
	setText(s.controls[setLabelConfirm], tr(lang, "confirm_label"))
	setText(s.controls[setAutoStart], tr(lang, "auto_start"))
	setText(s.controls[setStartMinimized], tr(lang, "start_minimized_tray"))
	setText(s.controls[setAutoMonitor], tr(lang, "auto_monitor"))
	setText(s.controls[setCloseToTray], tr(lang, "close_to_tray"))
	setText(s.controls[setAutoUpdate], "Automatically check GitHub for updates")
	setText(s.controls[setOutageNotify], "Show Windows notifications for outages and recovery")
	setText(s.controls[setInfo], "NetWatcher uses English for the interface, reports and log messages.")
	setText(s.controls[setSave], tr(lang, "save"))
	setText(s.controls[setCancel], tr(lang, "cancel"))
}
func (s *SettingsWindow) save() bool {
	interval, e1 := strconv.ParseFloat(strings.ReplaceAll(getText(s.controls[setInterval]), ",", "."), 64)
	timeout, e2 := strconv.Atoi(getText(s.controls[setTimeout]))
	latency, e3 := strconv.ParseFloat(strings.ReplaceAll(getText(s.controls[setLatency]), ",", "."), 64)
	confirm, e4 := strconv.Atoi(getText(s.controls[setConfirm]))
	lang := "en"
	theme := "light"
	if comboGet(s.controls[setTheme]) == 1 {
		theme = "dark"
	}
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || interval < 0.5 || timeout < 200 || latency < 1 || confirm < 1 {
		messageBoxOwned(s.hwnd, tr(lang, "settings_title"), tr(lang, "settings_invalid"), MB_OK|MB_ICONERROR)
		return false
	}
	a := s.parent
	a.mu.Lock()
	a.config.Language = "en"
	a.config.Theme = theme
	a.config.Interval = interval
	a.config.TimeoutMS = timeout
	a.config.HighLatencyMS = latency
	a.config.ConfirmCycles = confirm
	a.config.AutoStart = isChecked(s.controls[setAutoStart])
	a.config.StartMinimizedTray = isChecked(s.controls[setStartMinimized])
	a.config.AutoMonitor = isChecked(s.controls[setAutoMonitor])
	a.config.CloseToTray = isChecked(s.controls[setCloseToTray])
	a.config.AutoCheckUpdates = isChecked(s.controls[setAutoUpdate])
	a.config.OutageNotifications = isChecked(s.controls[setOutageNotify])
	cfg := a.config
	a.mu.Unlock()
	_ = saveConfig(cfg)
	a.applyAutoStart()
	procPostMessageW.Call(uintptr(a.hwnd), msgSyncTray, 0, 0)
	a.applyLanguage()
	a.applyTheme()
	messageBoxOwned(s.hwnd, tr(lang, "settings_title"), tr(lang, "settings_saved"), MB_OK|MB_ICONINFORMATION)
	return true
}
func settingsProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	s := globalSettings
	if s == nil {
		r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
		return r
	}
	switch msg {
	case WM_CREATE:
		s.hwnd = hwnd
		s.buildControls()
		applyWindowDarkMode(s.hwnd, isDarkTheme(s.parent.config.Theme))
		for _, h := range s.controls {
			applyControlTheme(h, isDarkTheme(s.parent.config.Theme))
		}
		return 0
	case WM_SIZE:
		s.layout(int32(loword(lParam)), int32(hiword(lParam)))
		return 0
	case WM_ERASEBKGND:
		fillClientBackground(hwnd, syscall.Handle(wParam), isDarkTheme(s.parent.config.Theme))
		return 1
	case WM_CTLCOLORSTATIC, WM_CTLCOLORBTN:
		return themeControlColor(syscall.Handle(wParam), isDarkTheme(s.parent.config.Theme), false)
	case WM_CTLCOLOREDIT:
		return themeControlColor(syscall.Handle(wParam), isDarkTheme(s.parent.config.Theme), true)
	case WM_COMMAND:
		id, notify := int(loword(wParam)), hiword(wParam)
		if id == setAutoStart && notify == BN_CLICKED {
			enable(s.controls[setStartMinimized], isChecked(s.controls[setAutoStart]))
			return 0
		}
		if id == setTheme && notify == CBN_SELCHANGE {
			theme := "light"
			if comboGet(s.controls[setTheme]) == 1 {
				theme = "dark"
			}
			s.parent.mu.Lock()
			s.parent.config.Theme = theme
			s.parent.mu.Unlock()
			s.parent.applyTheme()
			applyWindowDarkMode(s.hwnd, isDarkTheme(theme))
			for _, h := range s.controls {
				applyControlTheme(h, isDarkTheme(theme))
			}
			procInvalidateRect.Call(uintptr(s.hwnd), 0, 1)
			return 0
		}
		if notify == BN_CLICKED {
			if id == setSave {
				if s.save() {
					procDestroyWindow.Call(uintptr(hwnd))
				}
			} else if id == setCancel {
				s.parent.mu.Lock()
				s.parent.config.Language = "en"
				s.parent.config.Theme = s.originalTheme
				s.parent.mu.Unlock()
				s.parent.applyLanguage()
				s.parent.applyTheme()
				procDestroyWindow.Call(uintptr(hwnd))
			}
		}
		return 0
	case WM_CLOSE:
		s.parent.mu.Lock()
		s.parent.config.Language = "en"
		s.parent.config.Theme = s.originalTheme
		s.parent.mu.Unlock()
		s.parent.applyLanguage()
		s.parent.applyTheme()
		procDestroyWindow.Call(uintptr(hwnd))
		return 0
	case WM_DESTROY:
		enable(s.parent.hwnd, true)
		procSetForegroundWindow.Call(uintptr(s.parent.hwnd))
		globalSettings = nil
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

func windowProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	a := globalApp
	switch msg {
	case WM_CREATE:
		a.hwnd = hwnd
		a.buildControls()
		a.applyTheme()
		a.syncTrayIcon()
		a.refreshUI()
		if a.config.AutoCheckUpdates && strings.TrimSpace(githubRepository) != "" {
			a.checkForUpdates(false)
		}
		if a.config.AutoMonitor {
			go func() {
				time.Sleep(500 * time.Millisecond)
				a.startMonitoring(a.config.Interval, a.config.TimeoutMS)
				postRefresh(a.hwnd)
			}()
		}
		return 0
	case WM_SIZE:
		// Clicking the minimize button moves NetWatcher to the notification
		// area instead of leaving a taskbar button behind. Only hide the
		// window after Windows confirms that the tray icon was created; if
		// icon creation fails, fall back to ordinary minimization so the user
		// never loses access to the application.
		if a != nil && shouldMoveToTrayOnSize(wParam) {
			a.ensureTrayIcon()
			if a.trayAdded {
				a.startHidden = true
				procShowWindow.Call(uintptr(hwnd), SW_HIDE)
				return 0
			}
		}
		a.layout(int32(loword(lParam)), int32(hiword(lParam)))
		return 0
	case WM_ERASEBKGND:
		fillClientBackground(hwnd, syscall.Handle(wParam), isDarkTheme(a.config.Theme))
		return 1
	case WM_CTLCOLORSTATIC, WM_CTLCOLORBTN:
		return themeControlColor(syscall.Handle(wParam), isDarkTheme(a.config.Theme), false)
	case WM_CTLCOLOREDIT:
		return themeControlColor(syscall.Handle(wParam), isDarkTheme(a.config.Theme), true)
	case WM_COMMAND:
		if hiword(wParam) == BN_CLICKED {
			switch int(loword(wParam)) {
			case ctrlStart:
				a.mu.RLock()
				interval := a.config.Interval
				timeout := a.config.TimeoutMS
				a.mu.RUnlock()
				a.startMonitoring(interval, timeout)
				a.refreshUI()
			case ctrlStop:
				a.stopMonitoring(tr(a.config.Language, "user_stopped"))
			case ctrlAdd:
				if a.addCustomTarget(getText(a.controls[ctrlCustom])) {
					a.refreshCustomTargetCombo("")
				}
				a.refreshUI()
			case ctrlRemove:
				if a.removeCustomTarget(getText(a.controls[ctrlCustom])) {
					a.refreshCustomTargetCombo("")
				}
				a.refreshUI()
			case ctrlReport:
				path := a.report()
				if path == "" {
					messageBox(appName, tr(a.config.Language, "no_report"), MB_OK|MB_ICONINFORMATION)
				} else {
					_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
					a.addEvent(tr(a.config.Language, "report_created") + ": " + path)
					a.refreshUI()
				}
			case ctrlStats:
				path, err := generateStatisticsPage(a.logDir)
				if err != nil {
					messageBox(appName, "Statistics could not be generated:\n\n"+err.Error(), MB_OK|MB_ICONERROR)
				} else {
					_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
				}
			case ctrlExport:
				path, err := exportLogsZip(a.logDir)
				if err != nil {
					messageBox(appName, "Logs could not be exported:\n\n"+err.Error(), MB_OK|MB_ICONERROR)
				} else {
					a.addEvent("Log archive created: " + path)
					revealFile(path)
					a.refreshUI()
				}
			case ctrlLogs:
				_ = exec.Command("explorer", a.logDir).Start()
			case ctrlSettings:
				openSettings(a)
			}
		}
		return 0
	case msgRefresh:
		if a != nil {
			a.refreshUI()
		}
		return 0
	case msgTrayIcon:
		if a != nil {
			a.handleTrayMessage(uint32(lParam))
		}
		return 0
	case msgSyncTray:
		if a != nil {
			a.syncTrayIcon()
		}
		return 0
	case WM_PAINT:
		var ps PAINTSTRUCT
		hdc, _, _ := procBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
		var client RECT
		procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&client)))
		if a != nil {
			fillClientBackground(hwnd, syscall.Handle(hdc), isDarkTheme(a.config.Theme))
			a.drawGraph(syscall.Handle(hdc), client)
		}
		procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
		return 0
	case WM_CLOSE:
		if a != nil && a.config.CloseToTray && !a.exiting {
			a.ensureTrayIcon()
			procShowWindow.Call(uintptr(hwnd), SW_HIDE)
			return 0
		}
		if a != nil {
			a.exiting = true
			a.removeTrayIcon()
			a.stopMonitoring(tr(a.config.Language, "app_closed"))
		}
		procDestroyWindow.Call(uintptr(hwnd))
		return 0
	case WM_DESTROY:
		if a != nil {
			a.removeTrayIcon()
		}
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}
func runApp(startHidden bool) {
	globalApp = newApp()
	globalApp.startHidden = startHidden
	globalApp.runFirstLaunchQuestions()
	runWindow("NetWatcherNativeWindow", wndProcCallback, appName+" "+appVersion, 1120, 720, !startHidden)
}

func runWindow(class string, callback uintptr, title string, width, height int, showInitially bool) {
	style := uint32(WS_OVERLAPPEDWINDOW | WS_CLIPCHILDREN)
	if showInitially {
		style |= WS_VISIBLE
	}
	runWindowStyle(class, callback, title, width, height, style, showInitially)
}
func runWindowStyle(class string, callback uintptr, title string, width, height int, style uint32, showInitially bool) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	instance, _, _ := procGetModuleHandleW.Call(0)
	className := ptr(class)
	icon := loadIconFromFile(runtimeIconPath())
	if icon == 0 {
		fallback, _, _ := procLoadIconW.Call(0, IDI_APPLICATION)
		icon = syscall.Handle(fallback)
	}
	cursor, _, _ := procLoadCursorW.Call(0, IDC_ARROW)
	wc := WNDCLASSEX{CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), LpfnWndProc: callback, HInstance: syscall.Handle(instance), HIcon: icon, HCursor: syscall.Handle(cursor), HbrBackground: syscall.Handle(COLOR_WINDOW + 1), LpszClassName: className, HIconSm: icon}
	atom, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		if errno, ok := err.(syscall.Errno); !ok || errno != syscall.Errno(1410) {
			messageBox(appName, "Window class error: "+err.Error(), MB_OK|MB_ICONERROR)
			return
		}
	}
	hwnd, _, err := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(ptr(title))), uintptr(style), 120, 80, uintptr(width), uintptr(height), 0, 0, instance, 0)
	if hwnd == 0 {
		messageBox(appName, "Window error: "+err.Error(), MB_OK|MB_ICONERROR)
		return
	}
	procSendMessageW.Call(hwnd, WM_SETICON, ICON_BIG, uintptr(icon))
	procSendMessageW.Call(hwnd, WM_SETICON, ICON_SMALL, uintptr(icon))
	if showInitially {
		procShowWindow.Call(hwnd, SW_SHOW)
		procUpdateWindow.Call(hwnd)
	}
	var msg MSG
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "--app", "--portable":
			runApp(false)
			return
		case "--startup":
			runApp(true)
			return
		case "--uninstall":
			runUninstaller()
			return
		}
	}
	base := strings.ToLower(filepath.Base(currentExe()))
	if base == "netwatcher.exe" {
		runApp(false)
		return
	}
	runOneClickInstaller()
}
