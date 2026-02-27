package loader

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"log"
	"path"

	"github.com/hajimehoshi/ebiten/v2"
)

// AssetManager wraps an embed.FS and provides safe resource loading
// with placeholder fallback and image caching.
type AssetManager struct {
	fsys       fs.FS
	imageCache map[string]*ebiten.Image
}

// NewAssetManager creates an AssetManager backed by the given fs.FS.
func NewAssetManager(fsys fs.FS) *AssetManager {
	return &AssetManager{
		fsys:       fsys,
		imageCache: make(map[string]*ebiten.Image),
	}
}

// ReadFile reads raw bytes from the embedded filesystem.
func (am *AssetManager) ReadFile(name string) ([]byte, error) {
	clean := path.Clean(name)
	data, err := fs.ReadFile(am.fsys, clean)
	if err != nil {
		return nil, fmt.Errorf("loader: failed to read %s: %w", clean, err)
	}
	return data, nil
}

// LoadImage loads an image from the embedded filesystem with caching.
// On failure, returns a magenta/black checkerboard placeholder (never nil).
func (am *AssetManager) LoadImage(name string) *ebiten.Image {
	clean := path.Clean(name)

	// Return cached image if available
	if img, ok := am.imageCache[clean]; ok {
		return img
	}

	data, err := fs.ReadFile(am.fsys, clean)
	if err != nil {
		log.Printf("[WARN] loader: image not found %s: %v — using placeholder", clean, err)
		ph := generatePlaceholder()
		am.imageCache[clean] = ph
		return ph
	}

	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Printf("[WARN] loader: failed to decode image %s: %v — using placeholder", clean, err)
		ph := generatePlaceholder()
		am.imageCache[clean] = ph
		return ph
	}

	img := ebiten.NewImageFromImage(decoded)
	am.imageCache[clean] = img
	return img
}

// GenerateRect creates a solid-color rectangle image (Greyboxing protocol).
func GenerateRect(w, h int, c color.Color) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	return img
}

// generatePlaceholder creates a 64x64 magenta/black checkerboard image.
func generatePlaceholder() *ebiten.Image {
	const size = 64
	const cell = 8
	img := ebiten.NewImage(size, size)
	magenta := color.RGBA{R: 255, G: 0, B: 255, A: 255}
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if ((x/cell)+(y/cell))%2 == 0 {
				img.Set(x, y, magenta)
			} else {
				img.Set(x, y, black)
			}
		}
	}
	return img
}
