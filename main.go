package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/.output/public
var assets embed.FS

func main() {
	app := NewApp()
	sub, err := fs.Sub(assets, "frontend/.output/public")
	if err != nil {
		log.Fatal(err)
	}
	err = wails.Run(&options.App{
		Title:       "Photo Backup",
		Width:       1180,
		Height:      780,
		MinWidth:    920,
		MinHeight:   600,
		AssetServer: &assetserver.Options{Assets: sub},
		OnStartup:   app.startup,
		Bind:        []interface{}{app},
	})
	if err != nil {
		log.Fatal(err)
	}
}
