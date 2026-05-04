//go:build windows
// +build windows

package machineid

import (
	"golang.org/x/sys/windows/registry"
)

// readWindows is only compiled on Windows. It reads the canonical
// MachineGuid value from HKLM\SOFTWARE\Microsoft\Cryptography that
// Windows itself uses to identify the installation.
func readWindows() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Cryptography`,
		registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return "", err
	}
	defer k.Close()
	v, _, err := k.GetStringValue("MachineGuid")
	if err != nil {
		return "", err
	}
	return v, nil
}
