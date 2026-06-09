package main

import (
	"embed"
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/masahide/OmniSSHAgent/pkg/filelog"
	"github.com/masahide/OmniSSHAgent/pkg/pageant"
	"github.com/masahide/OmniSSHAgent/pkg/store"
	"github.com/masahide/OmniSSHAgent/pkg/store/local"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

const (
	AppName    = "OmniSSHAgent"
	AppVersion = "0.6.3"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var iconData []byte

func checkAlreadyRunning() {
	b, err := pageant.AlreadyRunning()
	if err != nil {
		return
	}
	//respLen := binary.BigEndian.Uint32(b[:4])
	if string(b[4:]) == AppName {
		os.Exit(0)
	}
}

var Logger *filelog.FileLog

func main() {
	// Tune Garbage Collector to run more aggressively to keep memory footprint minimal.
	debug.SetGCPercent(20)

	isService := flag.Bool("service", false, "run as a headless background service")
	isStartup := flag.Bool("startup", false, "run from Windows startup")
	port := flag.Int("port", 53210, "port for the HTTP API service")
	flag.Parse()

	if *isService {
		Logger = filelog.New(AppName, 1)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.SetOutput(Logger)

		checkAlreadyRunning()

		svc := NewService(*port)
		if err := svc.Start(); err != nil {
			log.Fatalf("failed to start service: %v", err)
		}
		select {} // Keep service running
	}

	Logger = filelog.New(AppName, 1)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(Logger)
	// Create an instance of the app structure

	checkAlreadyRunning()

	app := NewApp()
	app.settings = store.NewSettings(AppName, local.NewLocalCred(AppName))
	if err := app.settings.Load(); err != nil {
		log.Fatal(err.Error())
	}

	userCacheDir, err := os.UserCacheDir()
	if err == nil {
		userCacheDir = filepath.Join(userCacheDir, AppName)
	} else {
		log.Printf("cannot set user cache dir for Web View: %v", err)
		userCacheDir = ""
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:             AppName,
		Width:             900,
		Height:            900,
		MinWidth:          720,
		MinHeight:         570,
		MaxWidth:          1280,
		MaxHeight:         900,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       app.settings.StartHidden || *isStartup,
		HideWindowOnClose: true,
		// RGBA:              &options.RGBA{R: 33, G: 37, B: 43, A: 255},
		Assets:     assets,
		LogLevel:   logger.INFO,
		OnStartup:  app.startup,
		OnDomReady: app.domReady,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop: true,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewUserDataPath:  userCacheDir,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "My Application",
				Message: "",
				Icon:    iconData,
			},
		},
	})

	if err != nil {
		log.Print(err)
		app.Quit()
	}
}
