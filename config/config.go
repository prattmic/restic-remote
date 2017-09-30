package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// Dir returns the default system configuration directory for the named
// application.
func Dir(name string) (string, error) {
	if runtime.GOOS == "windows" {
		appData, ok := os.LookupEnv("APPDATA")
		if !ok {
			return "", fmt.Errorf("APPDATA not set")
		}

		return filepath.Join(appData, name), nil
	}

	xdg, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if ok {
		return filepath.Join(xdg, name), nil
	}

	home, ok := os.LookupEnv("HOME")
	if ok {
		return filepath.Join(home, ".config", name), nil
	}

	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("cannot determine current user: %v", err)
	}

	if u.HomeDir != "" {
		return filepath.Join(u.HomeDir, ".config", name), nil
	}

	return "", fmt.Errorf("unable to find config directory")
}
