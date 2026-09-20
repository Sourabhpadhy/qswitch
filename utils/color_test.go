package utils

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestDominantColorExtraction tests dominant color extraction against a synthetic test PNG
// containing transparent pixels, pure black background pixels, and dominant foreground pixels.
func TestDominantColorExtraction(t *testing.T) {
	// Create a synthetic 10x10 image:
	// - 50 pixels transparent (A = 0)
	// - 20 pixels pure black (R=0, G=0, B=0, A=255)
	// - 30 pixels dominant accent color #51DEBD (R=81, G=222, B=189, A=255)
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	accentColor := color.RGBA{R: 0x51, G: 0xDE, B: 0xBD, A: 0xFF}
	blackColor := color.RGBA{R: 0, G: 0, B: 0, A: 0xFF}
	transparentColor := color.RGBA{R: 0, G: 0, B: 0, A: 0}

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			idx := y*10 + x
			if idx < 50 {
				img.Set(x, y, transparentColor)
			} else if idx < 70 {
				img.Set(x, y, blackColor)
			} else {
				img.Set(x, y, accentColor)
			}
		}
	}

	// Write synthetic image to a temporary PNG file
	tmpFile, err := os.CreateTemp("", "synthetic_test_*.png")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := png.Encode(tmpFile, img); err != nil {
		tmpFile.Close()
		t.Fatalf("failed to encode png: %v", err)
	}
	tmpFile.Close()

	// Extract dominant color
	extracted := ExtractDominantColor(tmpFile.Name())
	expected := "#51DEBD"

	if extracted != expected {
		t.Errorf("expected dominant color %s, got %s", expected, extracted)
	}
}

// TestFallbackBehavior tests fallback to DefaultAccentColor when:
// 1. The image file does not exist.
// 2. An empty path is provided.
// 3. The image contains only transparent and pure black pixels.
// 4. GetFlavourColor is called for an unconfigured flavour.
func TestFallbackBehavior(t *testing.T) {
	// 1. Non-existent file path
	nonExistentPath := filepath.Join(os.TempDir(), "definitely_not_existing_asset_9999.png")
	colorNonExistent := ExtractDominantColor(nonExistentPath)
	if colorNonExistent != DefaultAccentColor {
		t.Errorf("expected fallback %s for non-existent file, got %s", DefaultAccentColor, colorNonExistent)
	}

	// 2. Empty image path
	colorEmpty := ExtractDominantColor("")
	if colorEmpty != DefaultAccentColor {
		t.Errorf("expected fallback %s for empty path, got %s", DefaultAccentColor, colorEmpty)
	}

	// 3. Image with only transparent and black pixels
	onlyBlackAndTransparent := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if (x+y)%2 == 0 {
				onlyBlackAndTransparent.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
			} else {
				onlyBlackAndTransparent.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
			}
		}
	}

	tmpFile, err := os.CreateTemp("", "black_transparent_*.png")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := png.Encode(tmpFile, onlyBlackAndTransparent); err != nil {
		tmpFile.Close()
		t.Fatalf("failed to encode png: %v", err)
	}
	tmpFile.Close()

	colorAllFiltered := ExtractDominantColor(tmpFile.Name())
	if colorAllFiltered != DefaultAccentColor {
		t.Errorf("expected fallback %s for all-filtered image, got %s", DefaultAccentColor, colorAllFiltered)
	}

	// 4. GetFlavourColor for a non-existent flavour
	flavourColor := GetFlavourColor("non_existent_flavour_xyz")
	if flavourColor != DefaultAccentColor {
		t.Errorf("expected fallback %s for unconfigured flavour, got %s", DefaultAccentColor, flavourColor)
	}
}

// TestAssetPathResolution tests path resolution for files inside ~/.config/qswitch/assets/.
func TestAssetPathResolution(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "qswitch_asset_home_*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	assetsDir := filepath.Join(tmpHome, ".config", "qswitch", "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatalf("failed to create assets dir: %v", err)
	}

	// Create an asset icon for "mycustomshell"
	customAssetPath := filepath.Join(assetsDir, "mycustomshell.png")
	if err := os.WriteFile(customAssetPath, []byte("fake png content"), 0644); err != nil {
		t.Fatalf("failed to write dummy asset: %v", err)
	}

	// Test resolution for existing asset
	resolved := GetAssetPath("mycustomshell")
	if resolved != customAssetPath {
		t.Errorf("expected resolved path %s, got %s", customAssetPath, resolved)
	}

	// Test resolution for non-existing asset
	resolvedMissing := GetAssetPath("nonexistent")
	if resolvedMissing != "" {
		t.Errorf("expected empty string for missing asset, got %s", resolvedMissing)
	}
}
