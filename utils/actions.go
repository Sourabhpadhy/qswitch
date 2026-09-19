package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func isQuickshellDir(name string) bool {
	roots := []string{
		filepath.Join(os.Getenv("HOME"), ".config", "quickshell"),
		"/etc/xdg/quickshell",
		"/usr/share/quickshell",
		"/usr/local/share/quickshell",
	}
	for _, root := range roots {
		p := filepath.Join(root, name)
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

func ApplyKeybinds(flavour string, config Config, debug bool) {
	// Handle keybinds
	qswitchDir := filepath.Join(os.Getenv("HOME"), ".config", "qswitch")
	qswitchCacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "qswitch")
	os.MkdirAll(qswitchDir, 0755)
	if err := os.MkdirAll(qswitchCacheDir, 0755); err != nil {
		fmt.Printf("Error creating cache directory: %v\n", err)
		return
	}
	keybindsFile := filepath.Join(qswitchCacheDir, "qswitch.lua")

	var contentParts []string

	flavourName := filepath.Base(flavour)

	contentParts = append(contentParts, "hl.bind(\""+config.PanelKeybind+"\", hl.dsp.exec_cmd(\"qswitch panel\"))")
	// Check for unbinds if enabled
	if config.Unbinds {
		unbindsPath := filepath.Join(
			os.Getenv("HOME"),
			".config",
			"qswitch",
			"keybinds",
			"unbinds.lua",
		)
		if info, err := os.Stat(unbindsPath); err == nil && !info.IsDir() {
			contentParts = append(contentParts, "dofile(\""+unbindsPath+"\")")
		} else {
			Debug(debug, "unbinds.lua not found at %s", unbindsPath)
		}
	}

	// Add mapped flavour keybinds dynamically
	keybindPath := GetKeybindPath(flavourName, config)
	if keybindPath != "" {
		contentParts = append(contentParts, "dofile(\""+keybindPath+"\")")
	} else if config.Keybinds[flavourName] == "default" {
		contentParts = append(contentParts, "-- Default")
	} else {
		fmt.Printf("Warning: keybind file not found for flavour %s\n", flavourName)
	}

	content := strings.Join(contentParts, "\n")
	Debug(debug, "ApplyKeybinds called with flavour='%s', extracted flavourName='%s'", flavour, flavourName)
	Debug(debug, "config.Keybinds[flavourName] is '%s'", config.Keybinds[flavourName])
	Debug(debug, "Final content to write:\n%s", content)
	if err := os.WriteFile(keybindsFile, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing keybinds file: %v\n", err)
	}
}

func ApplyFlavour(target string, config Config, debug bool) {
	flavourName := filepath.Base(target)

	Debug(debug, "Stopping previous instances (qs, caelestia, whisker)...")
	exec.Command("pkill", "-x", "qs").Run()
	exec.Command("caelestia", "shell", "-k").Run()
	exec.Command("whisker", "shell", "stop").Run()

	if flavourName == "dms" {
		Debug(debug, "Starting dms...")
		exec.Command("dms", "run", "-d").Run()
	} else if strings.ToLower(flavourName) == "ambxst" {
		Debug(debug, "Starting ambxst...")
		cmd := exec.Command("ambxst")
		cmd.Start()
	} else if flavourName == "whisker" {
		Debug(debug, "Starting whisker shell...")
		exec.Command("whisker", "shell").Run()
	} else if isQuickshellDir(flavourName) {
		Debug(debug, "Starting qs with target: %s...", flavourName)
		cmd := exec.Command("qs", "-c", flavourName)
		cmd.Start()
	} else if _, err := exec.LookPath(flavourName); err == nil {
		Debug(debug, "Starting %s directly...", flavourName)
		cmd := exec.Command(flavourName)
		cmd.Start()
	} else {
		Debug(debug, "Starting qs with fallback target: %s...", flavourName)
		cmd := exec.Command("qs", "-c", flavourName)
		cmd.Start()
	}

	Debug(debug, "Applying keybinds...")
	ApplyKeybinds(flavourName, config, debug)

	// Dynamically execute / source mapped lua file in Hyprland
	keybindPath := GetKeybindPath(flavourName, config)
	if keybindPath != "" {
		Debug(debug, "Dynamically executing mapped keybinds lua file: %s", keybindPath)
		exec.Command("hyprctl", "eval", fmt.Sprintf("dofile(\"%s\")", keybindPath)).Run()
	}

	Debug(debug, "Reloading hyprland...")
	exec.Command("hyprctl", "reload").Run()
}

// TogglePanel opens the panel if not running, closes it if running
func TogglePanel() {
	// Check if panel is already running by reading PID file
	pidData, err := os.ReadFile(panelPidFile)
	if err == nil {
		pid := strings.TrimSpace(string(pidData))
		// Check if process is still running
		checkCmd := exec.Command("kill", "-0", pid)
		if checkCmd.Run() == nil {
			// Process is running, kill it
			exec.Command("kill", pid).Run()
			os.Remove(panelPidFile)
			return
		}
	}

	// Panel not running, start it
	cmd := exec.Command("qs", "-c", "qswitch")
	cmd.Start()
	if cmd.Process != nil {
		os.WriteFile(panelPidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644)
	}
}

func Cycle(config Config) {
	current := ReadState()

	// Find the first installed flavour for fallback
	firstInstalled := ""
	for _, f := range config.Flavours {
		if IsFlavourInstalled(f, config) {
			firstInstalled = f
			break
		}
	}

	if firstInstalled == "" {
		fmt.Println("No installed flavours found.")
		return
	}

	if current == "" {
		WriteState(firstInstalled)
		ApplyFlavour(firstInstalled, config, false)
		fmt.Println("Switched to", firstInstalled)
		return
	}

	// Find current index and cycle to next installed flavour
	currentIdx := -1
	for i, f := range config.Flavours {
		if f == current {
			currentIdx = i
			break
		}
	}

	if currentIdx == -1 {
		// Current not found, use first installed
		WriteState(firstInstalled)
		ApplyFlavour(firstInstalled, config, false)
		fmt.Println("Switched to", firstInstalled)
		return
	}

	// Find next installed flavour
	for i := 1; i <= len(config.Flavours); i++ {
		nextIdx := (currentIdx + i) % len(config.Flavours)
		next := config.Flavours[nextIdx]
		if IsFlavourInstalled(next, config) {
			WriteState(next)
			ApplyFlavour(next, config, false)
			fmt.Println("Switched to", next)
			return
		}
	}

	fmt.Println("No other installed flavours to switch to.")
}

func Setup(config Config, force bool) {
	// Check if state file exists
	if _, err := os.Stat(stateFile); err == nil && !force {
		fmt.Println("Setup already completed (state file exists).")
		return
	}

	// Create state file if it doesn't exist
	if err := os.WriteFile(stateFile, []byte(""), 0644); err != nil {
		fmt.Printf("Error creating state file: %v\n", err)
	}

	// Create cache directory
	qswitchCacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "qswitch")
	if err := os.MkdirAll(qswitchCacheDir, 0755); err != nil {
		fmt.Printf("Error creating cache directory: %v\n", err)
		return
	}

	keybindsFile := filepath.Join(qswitchCacheDir, "qswitch.lua")
	content := "hl.bind(\"" + config.PanelKeybind + "\", hl.dsp.exec_cmd(\"qswitch panel\"))"
	if err := os.WriteFile(keybindsFile, []byte(content), 0644); err != nil {
		fmt.Printf("Error creating keybinds file: %v\n", err)
		return
	}
	hyprlandFile := filepath.Join(os.Getenv("HOME"), ".config", "hypr", "hyprland.lua")

	// Check if already sourced
	hyprContent, err := os.ReadFile(hyprlandFile)
	if err == nil {
		sourceLine := "dofile(\"" + qswitchCacheDir + "/qswitch.lua\")"
		if strings.Contains(string(hyprContent), sourceLine) {
			fmt.Println("Setup completed (already sourced)")
			return
		}
	}

	f, err := os.OpenFile(hyprlandFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening hyprland.lua:", err)
		return
	}
	defer f.Close()
	f.WriteString("\ndofile(\"" + qswitchCacheDir + "/qswitch.lua\")\n")
	fmt.Println("Setup completed")
}
