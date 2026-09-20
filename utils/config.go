package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
)

type Config struct {
	Flavours     []string          `json:"flavours"`
	Keybinds     map[string]string `json:"keybinds"`
	Unbinds      bool              `json:"unbinds"`
	PanelKeybind string            `json:"panel_keybind"`
}

var defaultConfig = Config{
	Flavours:     []string{},
	Keybinds:     map[string]string{},
	Unbinds:      false,
	PanelKeybind: "SUPER + ALT + P",
}

func LoadConfig() Config {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "qswitch")
	configPath := filepath.Join(configDir, "config.json")

	// Create dir if not exists
	os.MkdirAll(configDir, 0755)

	// Create keybinds dir
	keybindsDir := filepath.Join(configDir, "keybinds")
	os.MkdirAll(keybindsDir, 0755)

	// Create assets dir
	assetsDir := filepath.Join(configDir, "assets")
	os.MkdirAll(assetsDir, 0755)

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Create with default
		defaultData, _ := json.MarshalIndent(defaultConfig, "", "  ")
		os.WriteFile(configPath, defaultData, 0644)
		return defaultConfig
	}

	var config Config
	json.Unmarshal(data, &config)
	if config.Keybinds == nil {
		config.Keybinds = defaultConfig.Keybinds
		// Write back updated config
		updatedData, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile(configPath, updatedData, 0644)
	}

	if config.PanelKeybind == "" {
		config.PanelKeybind = defaultConfig.PanelKeybind
		// Write back updated config
		updatedData, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile(configPath, updatedData, 0644)
	}
	return config
}

func IsValidFlavour(name string, config Config) bool {
	if slices.Contains(config.Flavours, name) {
		return true
	}
	if _, ok := config.Keybinds[name]; ok {
		return true
	}
	return false
}

