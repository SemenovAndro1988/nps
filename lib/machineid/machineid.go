// Package machineid returns a stable identifier for the host the bot
// runs on. It mirrors the value that the Windows registry stores at
// HKLM\SOFTWARE\Microsoft\Cryptography\MachineGuid; on Linux it falls
// back to /etc/machine-id (or /var/lib/dbus/machine-id) and on macOS
// to ioreg's IOPlatformUUID.
//
// The value is a stable per-host identifier that does not depend on
// network configuration, which makes it the right thing to use as a
// bot identity: a single machine that reconnects from a different IP
// is still the same bot.
package machineid

import (
	"errors"
	"io/ioutil"
	"os/exec"
	"runtime"
	"strings"
)

// Get returns the host machine identifier or an error if it cannot
// be determined. The returned string is lower-case, contains no
// surrounding whitespace and is suitable for transport.
func Get() (string, error) {
	id, err := readPlatform()
	if err != nil {
		return "", err
	}
	id = strings.TrimSpace(id)
	id = strings.Trim(id, "\x00")
	id = strings.ToLower(id)
	if id == "" {
		return "", errors.New("empty machine identifier")
	}
	return id, nil
}

// readPlatform performs the OS-specific lookup. The non-Windows
// implementations live here because they are tiny; the Windows path
// (registry access) lives in machineid_windows.go behind a build tag
// so the rest of the codebase still compiles on non-Windows hosts.
func readPlatform() (string, error) {
	switch runtime.GOOS {
	case "linux":
		return readLinux()
	case "darwin":
		return readDarwin()
	case "windows":
		return readWindows()
	default:
		return readLinux()
	}
}

func readLinux() (string, error) {
	for _, p := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
		if b, err := ioutil.ReadFile(p); err == nil {
			s := strings.TrimSpace(string(b))
			if s != "" {
				return s, nil
			}
		}
	}
	return "", errors.New("no machine-id file found")
}

func readDarwin() (string, error) {
	out, err := exec.Command("/usr/sbin/ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "\"IOPlatformUUID\"") {
			continue
		}
		i := strings.LastIndex(line, "\"")
		if i <= 0 {
			continue
		}
		s := strings.Trim(line[strings.LastIndex(line[:i], "\"")+1:i], " \"")
		if s != "" {
			return s, nil
		}
	}
	return "", errors.New("IOPlatformUUID not found")
}
