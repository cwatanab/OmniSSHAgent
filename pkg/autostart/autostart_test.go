package autostart

import (
	"errors"
	"reflect"
	"testing"
)

func TestCreateTaskArgs(t *testing.T) {
	args, err := createTaskArgs(`C:\Program Files\OmniSSHAgent\OmniSSHAgent.exe`)
	if err != nil {
		t.Fatalf("createTaskArgs returned error: %v", err)
	}

	want := []string{
		"/Create",
		"/TN", TaskName,
		"/TR", `"C:\Program Files\OmniSSHAgent\OmniSSHAgent.exe" --startup`,
		"/SC", "ONLOGON",
		"/RL", "LIMITED",
		"/F",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args mismatch\nwant: %#v\n got: %#v", want, args)
	}
}

func TestCreateTaskArgsRejectsInvalidPath(t *testing.T) {
	tests := []string{
		"",
		`C:\bad"path\OmniSSHAgent.exe`,
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			if _, err := createTaskArgs(tt); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestEnableFallsBackToRunKeyWhenTaskRegistrationFails(t *testing.T) {
	restoreAutostartHooks(t)

	runSchtasks = func(args ...string) ([]byte, error) {
		return []byte("ERROR: Access is denied."), errors.New("exit status 1")
	}

	var gotExePath string
	enableRunKey = func(exePath string) error {
		gotExePath = exePath
		return nil
	}

	const exePath = `C:\Program Files\OmniSSHAgent\OmniSSHAgent.exe`
	if err := EnableForExecutable(exePath); err != nil {
		t.Fatalf("EnableForExecutable returned error: %v", err)
	}
	if gotExePath != exePath {
		t.Fatalf("fallback exe path mismatch: want %q, got %q", exePath, gotExePath)
	}
}

func TestDisableSkipsDeleteWhenTaskIsMissing(t *testing.T) {
	restoreAutostartHooks(t)

	var calls [][]string
	runSchtasks = func(args ...string) ([]byte, error) {
		calls = append(calls, args)
		return []byte("task not found"), errors.New("exit status 1")
	}
	isRunKeyEnabled = func() bool {
		return false
	}
	disableRunKey = func() error {
		return nil
	}

	if err := Disable(); err != nil {
		t.Fatalf("Disable returned error: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected only query call, got %d calls: %#v", len(calls), calls)
	}
}

func restoreAutostartHooks(t *testing.T) {
	t.Helper()
	oldRunSchtasks := runSchtasks
	oldEnableRunKey := enableRunKey
	oldDisableRunKey := disableRunKey
	oldIsRunKeyEnabled := isRunKeyEnabled
	t.Cleanup(func() {
		runSchtasks = oldRunSchtasks
		enableRunKey = oldEnableRunKey
		disableRunKey = oldDisableRunKey
		isRunKeyEnabled = oldIsRunKeyEnabled
	})
}
