package autostart

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const (
	TaskName   = "OmniSSHAgent"
	startupArg = "--startup"
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
)

var (
	runSchtasks     = defaultRunSchtasks
	enableRunKey    = defaultEnableRunKey
	disableRunKey   = defaultDisableRunKey
	isRunKeyEnabled = defaultIsRunKeyEnabled
)

func SetEnabled(enabled bool) error {
	if enabled {
		return Enable()
	}
	return Disable()
}

func Enable() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}
	return EnableForExecutable(exePath)
}

func EnableForExecutable(exePath string) error {
	args, err := createTaskArgs(exePath)
	if err != nil {
		return err
	}
	out, err := runSchtasks(args...)
	if err == nil {
		_ = disableRunKey()
		return nil
	}

	taskErr := commandError("register startup task", out, err)
	if runErr := enableRunKey(exePath); runErr != nil {
		return fmt.Errorf("%w; fallback registry startup failed: %w", taskErr, runErr)
	}

	return nil
}

func Disable() error {
	var taskErr error
	if isTaskEnabled() {
		out, err := runSchtasks("/Delete", "/TN", TaskName, "/F")
		if err != nil {
			taskErr = commandError("remove startup task", out, err)
		}
	}
	if err := disableRunKey(); err != nil {
		if taskErr != nil {
			return fmt.Errorf("%w; remove registry startup failed: %w", taskErr, err)
		}
		return fmt.Errorf("remove registry startup failed: %w", err)
	}
	return taskErr
}

func IsEnabled() bool {
	return isTaskEnabled() || isRunKeyEnabled()
}

func isTaskEnabled() bool {
	_, err := runSchtasks("/Query", "/TN", TaskName)
	return err == nil
}

func createTaskArgs(exePath string) ([]string, error) {
	command, err := startupCommand(exePath)
	if err != nil {
		return nil, err
	}
	return []string{
		"/Create",
		"/TN", TaskName,
		"/TR", command,
		"/SC", "ONLOGON",
		"/RL", "LIMITED",
		"/F",
	}, nil
}

func startupCommand(exePath string) (string, error) {
	if strings.TrimSpace(exePath) == "" {
		return "", errors.New("executable path is empty")
	}
	if strings.Contains(exePath, `"`) {
		return "", errors.New("executable path contains a double quote")
	}
	return `"` + exePath + `" ` + startupArg, nil
}

func defaultRunSchtasks(args ...string) ([]byte, error) {
	cmd := exec.Command("schtasks.exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.CombinedOutput()
}

func commandError(action string, out []byte, err error) error {
	message := strings.TrimSpace(string(out))
	if message == "" {
		return fmt.Errorf("%s: %w", action, err)
	}
	return fmt.Errorf("%s: %w: %s", action, err, message)
}

func defaultEnableRunKey(exePath string) error {
	command, err := startupCommand(exePath)
	if err != nil {
		return err
	}
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringValue(TaskName, command)
}

func defaultDisableRunKey() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err == registry.ErrNotExist {
		return nil
	}
	if err != nil {
		return err
	}
	defer k.Close()
	if err := k.DeleteValue(TaskName); err != nil && err != registry.ErrNotExist {
		return err
	}
	return nil
}

func defaultIsRunKeyEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(TaskName)
	return err == nil
}
