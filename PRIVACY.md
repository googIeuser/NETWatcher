# Privacy

NetWatcher is designed to work locally.

## Data stored on the computer

NetWatcher may store:

- application preferences under `%APPDATA%\NetWatcher`;
- daily connection samples and outage records under `Documents\NetWatcherLogs`;
- HTML reports and ZIP exports requested by the user.

These files are not uploaded by NetWatcher.

## Network requests made by the application

- Monitoring checks contact only the targets configured by the user and the built-in gateway, Cloudflare and Google targets.
- Automatic update checks contact GitHub's public Releases API when enabled.
- Experimental Access Mode sends DNS-over-HTTPS queries to Cloudflare and opens direct TCP connections to destinations requested by proxy-aware applications.

NetWatcher does not provide analytics, advertising, accounts or remote telemetry.

## Access Mode and browsing data

Access Mode is a local proxy running on `127.0.0.1`. It forwards traffic in memory and does not write requested URLs, page contents, cookies or credentials to the NetWatcher logs. HTTPS content remains encrypted between the application/browser and the destination server. The local proxy only fragments the initial encrypted stream while forwarding it.
