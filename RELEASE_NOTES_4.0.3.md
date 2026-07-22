# NetWatcher 4.0.3

NetWatcher 4.0.3 fixes Windows icon handling while keeping the existing application identity.

## Fixed

- Keeps the executable name as `netwatcher.exe`
- Keeps the product name as `NetWatcher`
- Forces the Windows runner resource to use the NetWatcher icon
- Rebuilds Windows resources without the stale Flutter icon cache
- Recreates desktop and Start Menu shortcuts
- Uses a dedicated NetWatcher shortcut icon
- Preserves tray mode, detailed outage history and report exports

## Downloads

- `NetWatcher_Setup_4.0.3.exe`
- `NetWatcher_4.0.3_Windows_Portable.zip`
- Matching `.sha256` files
