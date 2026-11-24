package assets

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"sync"

	"jetpack-joyride/assets"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	mu         sync.RWMutex
	images     = make(map[string]*ebiten.Image)
	fontBytes  = make(map[string][]byte)
	fontFaces  = make(map[string]font.Face)
	isLoaded   = false
)

// LoadAssets parses all embedded images and fonts from assets.FS.
func LoadAssets() error {
	mu.Lock()
	defer mu.Unlock()

	if isLoaded {
		return nil
	}

	entries, err := assets.FS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("reading assets dir: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		data, err := assets.FS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("reading asset %s: %w", name, err)
		}

		if len(name) > 4 && (name[len(name)-4:] == ".png" || name[len(name)-4:] == ".jpg") {
			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				return fmt.Errorf("decoding image %s: %w", name, err)
			}
			ebitenImg := ebiten.NewImageFromImage(img)
			images[name] = ebitenImg
			// Also register lowercase alias if exists (e.g. barry.png vs Barry.png)
			if name == "barry.png" {
				images["Barry.png"] = ebitenImg
			} else if name == "Barry.png" {
				images["barry.png"] = ebitenImg
			}
		} else if len(name) > 4 && name[len(name)-4:] == ".ttf" {
			fontBytes[name] = data
		}
	}

	isLoaded = true
	return nil
}

// GetImage returns the cached ebiten.Image for the given file name.
func GetImage(name string) *ebiten.Image {
	mu.RLock()
	defer mu.RUnlock()
	return images[name]
}

// GetFont returns an OpenType font.Face with the specified size (DPI = 72).
func GetFont(name string, size float64) font.Face {
	mu.Lock()
	defer mu.Unlock()

	key := fmt.Sprintf("%s_%f", name, size)
	if face, ok := fontFaces[key]; ok {
		return face
	}

	data, ok := fontBytes[name]
	if !ok {
		return nil
	}

	tt, err := opentype.Parse(data)
	if err != nil {
		return nil
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil
	}

	fontFaces[key] = face
	return face
}
