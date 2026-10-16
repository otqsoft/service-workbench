package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"service-workbench/internal/config"
	"service-workbench/internal/manager"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	exeDir := initExeDir()

	app := NewApp()
	app.exeDir = exeDir
	app.store = config.NewStore(exeDir)
	app.mgr = manager.New()

	err := wails.Run(&options.App{
		Title:  "Service Workbench",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 30, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
