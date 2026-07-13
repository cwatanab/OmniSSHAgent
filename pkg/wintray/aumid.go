package wintray

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	shell32                = windows.NewLazySystemDLL("shell32.dll")
	setCurrentProcessAUMID = shell32.NewProc("SetCurrentProcessExplicitAppUserModelID")
)

const AppUserModelID = "masahide.OmniSSHAgent"

func SetAppUserModelID() error {
	ptr, err := windows.UTF16PtrFromString(AppUserModelID)
	if err != nil {
		return err
	}
	r, _, _ := setCurrentProcessAUMID.Call(uintptr(unsafe.Pointer(ptr)))
	if r != 0 {
		return windows.Errno(r)
	}
	return nil
}
