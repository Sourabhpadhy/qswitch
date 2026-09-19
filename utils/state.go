package utils

import (
	"os"
	"path/filepath"
	"strings"
)

var stateFile = os.Getenv("HOME") + "/.switch_state"
var panelPidFile = os.Getenv("HOME") + "/.qswitch_panel_pid"

func ReadState() string {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func WriteState(f string) {
	os.WriteFile(stateFile, []byte(f), 0644)
}

// GetKeybindPath returns the absolute path to the mapped .lua keybind file for a flavour if it exists.
func GetKeybindPath(flavour string, config Config) string {
	keybindsDir := filepath.Join(os.Getenv("HOME"), ".config", "qswitch", "keybinds")

	// 1. Check if flavour has an explicit mapped file in config.Keybinds
	if mapped, ok := config.Keybinds[flavour]; ok && mapped != "" && mapped != "default" {
		targetPath := filepath.Join(keybindsDir, mapped)
		if info, err := os.Stat(targetPath); err == nil && !info.IsDir() {
			return targetPath
		}
		if !strings.HasSuffix(mapped, ".lua") {
			targetPathLua := filepath.Join(keybindsDir, mapped+".lua")
			if info, err := os.Stat(targetPathLua); err == nil && !info.IsDir() {
				return targetPathLua
			}
		}
	}

	// 2. Check fallback <flavour>.lua inside keybinds directory
	flavourLua := filepath.Join(keybindsDir, flavour+".lua")
	if info, err := os.Stat(flavourLua); err == nil && !info.IsDir() {
		return flavourLua
	}

	return ""
}

func GetFlavourPath(flavour string) (string, bool) {
	cfg := LoadConfig()
	path := GetKeybindPath(flavour, cfg)
	if path != "" {
		return path, true
	}
	return "", false
}

func IsFlavourInstalled(flavour string, config ...Config) bool {
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = LoadConfig()
	}
	return GetKeybindPath(flavour, cfg) != ""
}

func CheckFirstRun() bool {
	if _, err := os.Stat(stateFile); err == nil {
		return false
	}
	return true
}