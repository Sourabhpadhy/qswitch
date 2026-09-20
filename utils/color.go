package utils

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// DefaultAccentColor is the fallback accent color when no icon or valid pixels exist.
const DefaultAccentColor = "#b4befe"

// GetAssetPath returns the absolute path to ~/.config/qswitch/assets/<flavour>.png if it exists.
func GetAssetPath(flavour string) string {
	home := os.Getenv("HOME")
	if home == "" {
		return ""
	}
	assetPath := filepath.Join(home, ".config", "qswitch", "assets", flavour+".png")
	if info, err := os.Stat(assetPath); err == nil && !info.IsDir() {
		return assetPath
	}
	return ""
}

// ExtractDominantColor reads a PNG file, filters out transparent or pure black background pixels,
// and extracts the dominant/accent color as a hex code (#RRGGBB).
// If the file does not exist, cannot be decoded, or contains only filtered pixels, it falls back to DefaultAccentColor.
func ExtractDominantColor(imagePath string) string {
	if imagePath == "" {
		return DefaultAccentColor
	}

	file, err := os.Open(imagePath)
	if err != nil {
		return DefaultAccentColor
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return DefaultAccentColor
	}

	return ExtractDominantColorFromImage(img)
}

// ExtractDominantColorFromImage analyzes an image.Image, filters out transparent
// and pure black pixels, and extracts the dominant/accent color as #RRGGBB.
func ExtractDominantColorFromImage(img image.Image) string {
	bounds := img.Bounds()
	if bounds.Empty() {
		return DefaultAccentColor
	}

	type bucketInfo struct {
		exactCounts map[uint32]int
		totalCount  int
	}

	bucketMap := make(map[uint32]*bucketInfo)
	totalValidPixels := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// Filter out transparent pixels (alpha below ~12% opacity)
			if a < 0x2000 {
				continue
			}

			// Convert 16-bit premultiplied RGBA to 8-bit un-premultiplied values
			r8 := uint8((r * 0xFF) / a)
			g8 := uint8((g * 0xFF) / a)
			b8 := uint8((b * 0xFF) / a)

			// Filter out pure black background pixels
			if r8 == 0 && g8 == 0 && b8 == 0 {
				continue
			}

			exact := (uint32(r8) << 16) | (uint32(g8) << 8) | uint32(b8)
			// Cluster into color buckets (quantize each channel to 4-bit / 16 shades)
			bucketKey := ((uint32(r8) >> 4) << 8) | ((uint32(g8) >> 4) << 4) | (uint32(b8) >> 4)

			bInfo, exists := bucketMap[bucketKey]
			if !exists {
				bInfo = &bucketInfo{exactCounts: make(map[uint32]int)}
				bucketMap[bucketKey] = bInfo
			}
			bInfo.exactCounts[exact]++
			bInfo.totalCount++
			totalValidPixels++
		}
	}

	if totalValidPixels == 0 {
		return DefaultAccentColor
	}

	// Find the bucket with the highest count
	var bestBucket *bucketInfo
	maxBucketCount := -1
	for _, bInfo := range bucketMap {
		if bInfo.totalCount > maxBucketCount {
			maxBucketCount = bInfo.totalCount
			bestBucket = bInfo
		}
	}

	if bestBucket == nil {
		return DefaultAccentColor
	}

	// Within the dominant bucket, select the most frequent exact color
	var bestExact uint32
	maxExactCount := -1
	for color, count := range bestBucket.exactCounts {
		if count > maxExactCount {
			maxExactCount = count
			bestExact = color
		}
	}

	return fmt.Sprintf("#%02X%02X%02X", uint8(bestExact>>16), uint8(bestExact>>8), uint8(bestExact))
}

// GetFlavourColor returns the dominant color for a flavour by inspecting its asset PNG,
// or returns DefaultAccentColor if no image exists or extraction yields no accent color.
func GetFlavourColor(flavour string) string {
	assetPath := GetAssetPath(flavour)
	if assetPath != "" {
		return ExtractDominantColor(assetPath)
	}
	return DefaultAccentColor
}
