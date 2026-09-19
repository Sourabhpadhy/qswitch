package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidFlavour(t *testing.T) {
	cfg := Config{
		Flavours: []string{"custom-shell-1", "ii"},
		Keybinds: map[string]string{
			"custom-shell-2": "custom2.lua",
		},
	}

	if !IsValidFlavour("custom-shell-1", cfg) {
		t.Errorf("expected custom-shell-1 to be valid")
	}
	if !IsValidFlavour("custom-shell-2", cfg) {
		t.Errorf("expected custom-shell-2 to be valid")
	}
	if IsValidFlavour("unknown-shell", cfg) {
		t.Errorf("expected unknown-shell to be invalid")
	}
}

func TestIsFlavourInstalled(t *testing.T) {
	// Create temporary directory structure imitating ~/.config/qswitch/keybinds
	tmpHome, err := os.MkdirTemp("", "qswitch_test_home")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	keybindsDir := filepath.Join(tmpHome, ".config", "qswitch", "keybinds")
	if err := os.MkdirAll(keybindsDir, 0755); err != nil {
		t.Fatalf("failed to create keybinds dir: %v", err)
	}

	// Create a mapped keybind file and an unmapped default keybind file
	if err := os.WriteFile(filepath.Join(keybindsDir, "mapped_file.lua"), []byte("-- test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(keybindsDir, "fallback.lua"), []byte("-- test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg := Config{
		Flavours: []string{"myflavour", "fallback", "missing"},
		Keybinds: map[string]string{
			"myflavour": "mapped_file.lua",
			"missing":   "notfound.lua",
		},
	}

	if !IsFlavourInstalled("myflavour", cfg) {
		t.Errorf("expected myflavour to be installed")
	}
	if !IsFlavourInstalled("fallback", cfg) {
		t.Errorf("expected fallback to be installed")
	}
	if IsFlavourInstalled("missing", cfg) {
		t.Errorf("expected missing to NOT be installed")
	}
}
