package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	if err := wails.Run(&options.App{
		Title:            "NetWatcher",
		Width:            1280,
		Height:           800,
		MinWidth:         920,
		MinHeight:        620,
		DisableResize:    false,
		BackgroundColour: &options.RGBA{R: 14, G: 18, B: 27, A: 1},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "cf5bd438-2b6d-45e5-8671-6e38f26b84e8",
			OnSecondInstanceLaunch: func(options.SecondInstanceData) {
				app.ShowWindow()
			},
		},
		Windows: &windows.Options{
			Theme:                windows.SystemDefault,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisablePinchZoom:     true,
		},
	}); err != nil {
		log.Fatal(err)
	}
}
