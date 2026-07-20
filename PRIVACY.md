# Privacy

NetWatcher is designed to work locally.

## Data stored on the computer

NetWatcher stores configuration under the current user's roaming application-data directory and writes CSV/HTML diagnostic files to `Documents\NetWatcherLogs`.

## Network access

NetWatcher sends ICMP echo requests to the configured monitoring targets. When automatic update checks are enabled, it requests the latest public release metadata from the GitHub API for this repository.

## Data not collected

NetWatcher does not include telemetry, analytics, advertising, user accounts, cloud synchronization, or automatic log uploads. It does not collect browsing history or inspect network payloads.
