package wintray

import (
	"log"

	"golang.org/x/sys/windows/registry"
)

func EnsureRegistered() {
	_ = SetAppUserModelID()

	k, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Classes\AppUserModelId\`+AppUserModelID,
		registry.SET_VALUE|registry.CREATE_SUB_KEY)
	if err != nil {
		k, _, err = registry.CreateKey(registry.CURRENT_USER,
			`Software\Classes\AppUserModelId\`+AppUserModelID,
			registry.SET_VALUE)
		if err != nil {
			log.Printf("wintray: failed to create AUMID registry key: %v", err)
			return
		}
	}
	defer k.Close()
	_ = k.SetStringValue("DisplayName", "OmniSSHAgent")
}

// SetRegisteredIconPath sets the IconUri value for the AUMID registry key
// so that Windows toast notifications display the app icon in the title bar.
func SetRegisteredIconPath(iconPath string) {
	k, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Classes\AppUserModelId\`+AppUserModelID,
		registry.SET_VALUE)
	if err != nil {
		log.Printf("wintray: failed to open AUMID registry key for IconUri: %v", err)
		return
	}
	defer k.Close()
	if err := k.SetExpandStringValue("IconUri", iconPath); err != nil {
		log.Printf("wintray: failed to set IconUri: %v", err)
	}
}
