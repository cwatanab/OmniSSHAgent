package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	rtdebug "runtime/debug"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/sys/windows/registry"

	"github.com/masahide/OmniSSHAgent/pkg/cygwinsocket"
	"github.com/masahide/OmniSSHAgent/pkg/namedpipe"
	"github.com/masahide/OmniSSHAgent/pkg/pageant"
	"github.com/masahide/OmniSSHAgent/pkg/sshkey"
	"github.com/masahide/OmniSSHAgent/pkg/sshutil"
	"github.com/masahide/OmniSSHAgent/pkg/store"
	"github.com/masahide/OmniSSHAgent/pkg/unix"
	"github.com/masahide/OmniSSHAgent/pkg/winopen"
	"github.com/masahide/OmniSSHAgent/pkg/wintray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App application struct
type App struct {
	ctx              context.Context
	backgroundCtx    context.Context
	cancelBackground context.CancelFunc
	agentCtx         context.Context
	ti               *wintray.TrayIcon
	keyRing          *sshutil.KeyRing
	settings         *store.Settings
	wg               sync.WaitGroup
	agentWG          sync.WaitGroup
	agentsMu         sync.Mutex
	cancelAgents     context.CancelFunc
	pageantCheckFunc func()
	shutdownOnce     sync.Once
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// trayStrings holds localised strings for the system tray menu.
type trayStrings struct {
	TooltipFmt    string // fmt string: AppName + key count
	ShowWindow    string
	ShowWindowTip string
	Quit          string
	QuitTip       string
	KeyUsed       string // fmt string: key name
}

var trayEN = trayStrings{
	TooltipFmt:    "%s - %d keys loaded",
	ShowWindow:    "Show Window",
	ShowWindowTip: "Show main window",
	Quit:          "Quit",
	QuitTip:       "Quit the whole app",
	KeyUsed:       "SSH Key '%s' was used",
}

var trayJA = trayStrings{
	TooltipFmt:    "%s - 鍵 %d 個読み込み済み",
	ShowWindow:    "ウィンドウを表示",
	ShowWindowTip: "メインウィンドウを表示します",
	Quit:          "終了",
	QuitTip:       "アプリケーションを終了します",
	KeyUsed:       "SSH鍵 '%s' が使用されました",
}

// getSystemLang returns "ja" when the Windows UI locale starts with "ja", otherwise "en".
func getSystemLang() string {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Control Panel\International`, registry.QUERY_VALUE)
	if err != nil {
		return "en"
	}
	defer k.Close()
	locale, _, err := k.GetStringValue("LocaleName")
	if err != nil {
		return "en"
	}
	if len(locale) >= 2 && locale[:2] == "ja" {
		return "ja"
	}
	return "en"
}

func trayStr() trayStrings {
	if getSystemLang() == "ja" {
		return trayJA
	}
	return trayEN
}

func (a *App) setTrayTooltip() {
	if a.ti == nil {
		return
	}
	ts := trayStr()
	tooltip := AppName
	if a.keyRing != nil {
		keys, err := a.keyRing.KeyList()
		if err == nil {
			tooltip = fmt.Sprintf(ts.TooltipFmt, AppName, len(keys))
		}
	}
	a.ti.SetTooltip(tooltip)
}

func (a *App) systrayOnReady() {
	// Systray operations must be executed on a dedicated tray thread (see doc/dev/issue-tasktry.md).
	ts := trayStr()
	a.ti.SetTitle(AppName)
	a.setTrayTooltip()
	mShowWindow := a.ti.AddMenuItem(ts.ShowWindow, ts.ShowWindowTip)
	mQuit := a.ti.AddMenuItem(ts.Quit, ts.QuitTip)
	go func() {
		for {
			select {
			case <-mShowWindow.ClickedCh:
				a.showWindow()
			case <-mQuit.ClickedCh:
				a.Quit()
				return
			}
		}
	}()
}
func (a *App) systrayOnExit() {
	//log.Print("systrayOnExit")
	Logger.Close()
	a.wg.Done()
}

func (a *App) initializeKeyRing() {
	a.keyRing = sshutil.NewKeyRing(a.settings)
	if err := a.keyRing.AddKeys(); err != nil {
		log.Printf("KeyRing.AddKeys err: %s", err)
	}
	a.keyRing.NotifyCallback = a.notice
}

func (a *App) startConfiguredAgents() {
	a.agentsMu.Lock()
	defer a.agentsMu.Unlock()
	a.startConfiguredAgentsLocked()
}

func (a *App) restartConfiguredAgents(rebuildKeyRing bool) {
	a.agentsMu.Lock()
	defer a.agentsMu.Unlock()
	a.stopConfiguredAgentsLocked()
	if rebuildKeyRing {
		a.initializeKeyRing()
	}
	a.startConfiguredAgentsLocked()
}

func (a *App) stopConfiguredAgents() {
	a.agentsMu.Lock()
	defer a.agentsMu.Unlock()
	a.stopConfiguredAgentsLocked()
}

func (a *App) stopConfiguredAgentsLocked() {
	if a.cancelAgents != nil {
		a.cancelAgents()
		a.cancelAgents = nil
	}
	a.agentWG.Wait()
	a.agentCtx = nil
}

func (a *App) startConfiguredAgentsLocked() {
	if a.settings == nil || a.keyRing == nil {
		return
	}
	agentCtx, cancel := context.WithCancel(context.Background())
	a.agentCtx = agentCtx
	a.cancelAgents = cancel

	debug := false
	checkFunc := a.pageantCheckFunc
	if checkFunc == nil {
		checkFunc = func() {}
	}

	if a.settings.PageantAgent {
		pa := &pageant.Pageant{
			ExtendedAgent: a.keyRing,
			AppName:       AppName,
			Debug:         debug,
			CheckFunc:     checkFunc,
		}
		log.Println("Starting pageant...")
		a.agentWG.Add(1)
		go func() {
			defer a.agentWG.Done()
			pa.RunAgent(agentCtx)
		}()

		pageantPipeName, err := pageant.ObfuscatedPipeName()
		if err != nil {
			log.Printf("Failed to compute Pageant named pipe name: %v", err)
		} else {
			paPipe := &namedpipe.NamedPipe{ExtendedAgent: a.keyRing, Debug: debug, Name: pageantPipeName}
			log.Printf("Starting Pageant named pipe: %s", pageantPipeName)
			a.agentWG.Add(1)
			go func() {
				defer a.agentWG.Done()
				if err := paPipe.RunAgent(agentCtx); err != nil {
					log.Printf("Pageant named pipe agent error: %v", err)
				}
			}()
		}
	}
	if a.settings.NamedPipeAgent {
		pipeName := ""
		na := &namedpipe.NamedPipe{ExtendedAgent: a.keyRing, Debug: debug, Name: pipeName}
		log.Println("Starting NamedPipe agent..")
		a.agentWG.Add(1)
		go func() {
			defer a.agentWG.Done()
			if err := na.RunAgent(agentCtx); err != nil {
				log.Printf("NamedPipe agent error: %v", err)
			}
		}()
	}
	if a.settings.UnixSocketAgent {
		ua := &unix.DomainSock{ExtendedAgent: a.keyRing, Debug: debug, Path: a.settings.UnixSocketPath}
		log.Println("Start Unix domain socket agent..")
		a.agentWG.Add(1)
		go func() {
			defer a.agentWG.Done()
			if err := ua.RunAgent(agentCtx); err != nil {
				log.Printf("Unix socket agent error: %v", err)
			}
		}()
	}
	if a.settings.CygWinAgent {
		ca := &cygwinsocket.CygwinSock{ExtendedAgent: a.keyRing, Debug: debug, Path: a.settings.CygWinSocketPath}
		log.Println("Starting Cygwin unix domain socket agent..")
		a.agentWG.Add(1)
		go func() {
			defer a.agentWG.Done()
			if err := ca.RunAgent(agentCtx); err != nil {
				log.Printf("Cygwin socket agent error: %v", err)
			}
		}()
	}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx
	a.backgroundCtx, a.cancelBackground = context.WithCancel(context.Background())
	a.pageantCheckFunc = a.showWindow
	if a.settings != nil {
		Logger.SetEnable(a.settings.SaveData.DebugLog)
	}
	a.ti = wintray.NewTrayIcon()
	a.ti.BalloonClickFunc = a.showWindow
	a.ti.TrayClickFunc = a.showWindow

	a.wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("recover:%v", r)
			}
			a.Quit()
		}()
		a.ti.Run(a.systrayOnReady, a.systrayOnExit)
	}()

	a.initializeKeyRing()
	a.startConfiguredAgents()

	// Spawn a background reclaimer to return unused memory to OS periodically
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-a.backgroundCtx.Done():
				return
			case <-ticker.C:
				rtdebug.FreeOSMemory()
			}
		}
	}()
}

func (a *App) notice(action string, data interface{}) {
	switch action {
	case "Add", "Remove", "RemoveAll":
		//a.ti.ShowBalloonNotification(action, sshutil.JSONDump(data))
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "LoadKeysEvent")
		}

	case "Added", "Removed", "RemovedAll":
		if a.ti != nil {
			a.setTrayTooltip()
		}

	case "Sign", "SignWithFlags":
		switch t := data.(type) {
		case *agent.Key:
			if err := a.onSign(t); err != nil {
				log.Printf("cannot find key to print: %v\n", err)
			}
		case ssh.PublicKey:
			log.Printf("unexpected ssh.PublicKey\n")
		}
	}
	// Async reclaim memory after key operations
	go rtdebug.FreeOSMemory()
}

func (a *App) onSign(pubkey *agent.Key) error {
	privkey := a.keyRing.FindPrivKey(pubkey)
	if privkey == nil {
		return errors.New("private key not found")
	}

	name := privkey.PublicKey.Comment
	if len(name) == 0 {
		name = truncateString(privkey.PublicKey.SHA256)
	}
	ts := trayStr()
	msg := fmt.Sprintf(ts.KeyUsed, name)
	if a.settings.ShowBalloon && a.ti != nil {
		a.ti.ShowBalloonNotification(wintray.ID, msg)
	}
	return nil
}

func (a *App) OpenFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select a private key file",
	})
}

// domReady is called after the front-end dom has been loaded
func (a *App) domReady(ctx context.Context) {
	// Free memory loaded during frontend startup
	go func() {
		time.Sleep(2 * time.Second) // wait for WebView to settle
		rtdebug.FreeOSMemory()
	}()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s!", name)
}

func (a *App) GetVersion() string {
	return AppVersion
}

func (a *App) showWindow() {
	//runtime.LogDebug(a.ctx, "showWindow")
	runtime.WindowShow(a.ctx)
}

func (a *App) Quit() {
	log.Print("call a.Quit")
	runtime.Quit(a.ctx)
}

// shutdown はruntime.Quitから呼ばれる
func (a *App) shutdown(ctx context.Context) {
	a.shutdownOnce.Do(func() {
		log.Print("shutdown")
		a.stopConfiguredAgents()
		if a.cancelBackground != nil {
			a.cancelBackground()
		}
		if a.ti != nil {
			a.ti.Quit()
		}
		a.wg.Wait()
	})
}

func (a *App) setDebugLogEnabled(enabled bool) {
	Logger.SetEnable(enabled)
	if a.settings != nil {
		a.settings.SaveData.DebugLog = enabled
	}
}

// OpenLogDir opens the directory containing the log file in Windows Explorer
func (a *App) OpenLogDir() {
	confDir, err := os.UserConfigDir()
	if err != nil {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Error",
			Message: fmt.Sprintf("Failed to locate config directory: %v", err),
		})
		return
	}
	dir := filepath.Join(confDir, AppName, "logs")
	if Logger != nil && Logger.FilePath != "" {
		dir = filepath.Dir(Logger.FilePath)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Error",
			Message: fmt.Sprintf("Failed to create log directory: %v", err),
		})
		return
	}
	if err := winopen.Open(dir); err != nil {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Error",
			Message: fmt.Sprintf("Failed to open directory: %v", err),
		})
	}
}

func (a *App) AddLocalFile(pk sshkey.PrivateKeyFile) error {
	pk.Name = filepath.Base(pk.FilePath)
	pk.StoreType = sshutil.LocalStore
	//log.Printf("AddLocalFile:%s", sshutil.JSONDump(pk))
	id, err := a.keyRing.AddKeySettings(pk)
	if err != nil {
		return err
	}
	if err := a.keyRing.AddKey(id); err != nil {
		return err
	}
	return nil
}

func (a *App) DeleteKey(sha256 string) error {
	if err := a.keyRing.RemoveKey(sha256); err != nil {
		return err
	}
	return a.keyRing.DeleteKeySettings(sha256)
}

func (a *App) ToggleKey(sha256 string) error {
	return a.keyRing.ToggleKey(sha256)
}

func (a *App) KeyList() ([]sshkey.PrivateKeyFile, error) {
	return a.keyRing.KeyList()
}

func (a *App) CheckKeyType(filePath, passphrase string) (*sshkey.PrivateKeyFile, error) {
	return sshutil.CheckKeyType(filePath, passphrase)
}

func (a *App) GetSettings() store.SaveData {
	return a.settings.SaveData
}
func (a *App) Save(s store.SaveData) error {
	old := a.settings.SaveData
	a.setDebugLogEnabled(s.DebugLog)
	a.settings.SaveData.StartHidden = s.StartHidden
	a.settings.SaveData.PageantAgent = s.PageantAgent
	a.settings.SaveData.NamedPipeAgent = s.NamedPipeAgent
	a.settings.SaveData.UnixSocketAgent = s.UnixSocketAgent
	a.settings.SaveData.UnixSocketPath = s.UnixSocketPath
	a.settings.SaveData.CygWinAgent = s.CygWinAgent
	a.settings.SaveData.ShowBalloon = s.ShowBalloon
	a.settings.SaveData.CygWinSocketPath = s.CygWinSocketPath
	a.settings.SaveData.ProxyModeOfNamedPipe = s.ProxyModeOfNamedPipe
	if err := a.settings.Save(); err != nil {
		a.settings.SaveData = old
		a.setDebugLogEnabled(old.DebugLog)
		return err
	}
	if agentSettingsChanged(old, a.settings.SaveData) {
		a.restartConfiguredAgents(old.ProxyModeOfNamedPipe != a.settings.ProxyModeOfNamedPipe)
	}
	return nil
}

func agentSettingsChanged(old, current store.SaveData) bool {
	return old.PageantAgent != current.PageantAgent ||
		old.NamedPipeAgent != current.NamedPipeAgent ||
		old.UnixSocketAgent != current.UnixSocketAgent ||
		old.UnixSocketPath != current.UnixSocketPath ||
		old.CygWinAgent != current.CygWinAgent ||
		old.CygWinSocketPath != current.CygWinSocketPath ||
		old.ProxyModeOfNamedPipe != current.ProxyModeOfNamedPipe
}

func truncateString(s string) string {
	if len(s) <= 16 {
		return s
	}
	return s[:16] + "..."
}

// GetAccentColor returns the Windows accent color as a hex string (e.g. "#4a5459")
func (a *App) GetAccentColor() string {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\DWM`, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue("AccentColor")
	if err != nil {
		return ""
	}

	// val is in AABBGGRR format.
	r := byte(val & 0xFF)
	g := byte((val >> 8) & 0xFF)
	b := byte((val >> 16) & 0xFF)

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// GetAppsUseLightTheme returns true if Windows is configured to use Light Theme for apps
func (a *App) GetAppsUseLightTheme() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return false
	}
	return val != 0
}
