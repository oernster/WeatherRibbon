//go:build linux || darwin

package main

import _ "embed"

// trayIcon is the application's icon, which the tray on Linux and the menu bar on macOS are handed
// as an image; on Windows the tray reads the icon built into the executable instead.
//
//go:embed build/appicon.png
var trayIcon []byte
