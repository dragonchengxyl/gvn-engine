package render

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	VirtualWidth  = 1920
	VirtualHeight = 1080
)

// CharPosition defines preset positions for character sprites.
type CharPosition string

const (
	PosLeft   CharPosition = "left"
	PosCenter CharPosition = "center"
	PosRight  CharPosition = "right"
)

// CharSprite holds a character's image and display position.
type CharSprite struct {
	Name     string
	Image    *ebiten.Image
	Position CharPosition
	Alpha    float64 // 0.0 ~ 1.0
}

// Renderer manages multi-layer drawing: BG → Characters → Foreground → UI.
type Renderer struct {
	bgImage    *ebiten.Image
	bgAlpha    float64
	chars      []*CharSprite
	fgOverlay  *ebiten.Image // reserved for effects (rain, etc.)
	fgAlpha    float64

	// Transition state
	oldBG         *ebiten.Image
	transProgress float64 // 0.0 = old scene, 1.0 = new scene
	transActive   bool
	transDuration float64 // in seconds
	transElapsed  float64
}

// NewRenderer creates a Renderer with default values.
func NewRenderer() *Renderer {
	return &Renderer{
		bgAlpha: 1.0,
		fgAlpha: 0.0,
	}
}

// SetBackground sets the background image, optionally with a fade transition.
func (r *Renderer) SetBackground(img *ebiten.Image, fadeDuration float64) {
	if fadeDuration > 0 && r.bgImage != nil {
		r.oldBG = r.bgImage
		r.transActive = true
		r.transDuration = fadeDuration
		r.transElapsed = 0
		r.transProgress = 0
	}
	r.bgImage = img
	r.bgAlpha = 1.0
}

// SetCharacter adds or updates a character sprite.
func (r *Renderer) SetCharacter(name string, img *ebiten.Image, pos CharPosition) {
	// Update existing
	for _, cs := range r.chars {
		if cs.Name == name {
			cs.Image = img
			cs.Position = pos
			cs.Alpha = 1.0
			return
		}
	}
	// Add new
	r.chars = append(r.chars, &CharSprite{
		Name:     name,
		Image:    img,
		Position: pos,
		Alpha:    1.0,
	})
}

// RemoveCharacter removes a character sprite by name.
func (r *Renderer) RemoveCharacter(name string) {
	for i, cs := range r.chars {
		if cs.Name == name {
			r.chars = append(r.chars[:i], r.chars[i+1:]...)
			return
		}
	}
	log.Printf("[WARN] render: character %q not found for removal", name)
}

// ClearCharacters removes all character sprites.
func (r *Renderer) ClearCharacters() {
	r.chars = r.chars[:0]
}

// SetForeground sets the foreground overlay image with a given alpha.
func (r *Renderer) SetForeground(img *ebiten.Image, alpha float64) {
	r.fgOverlay = img
	r.fgAlpha = alpha
}

// ClearForeground removes the foreground overlay.
func (r *Renderer) ClearForeground() {
	r.fgOverlay = nil
	r.fgAlpha = 0
}

// IsTransitioning returns true if a fade transition is in progress.
func (r *Renderer) IsTransitioning() bool {
	return r.transActive
}

// Update advances transition animations. Call once per tick (60 TPS).
func (r *Renderer) Update() {
	if !r.transActive {
		return
	}
	r.transElapsed += 1.0 / 60.0
	r.transProgress = r.transElapsed / r.transDuration
	if r.transProgress >= 1.0 {
		r.transProgress = 1.0
		r.transActive = false
		r.oldBG = nil
	}
}

// Draw renders all layers onto the screen. Pure function of renderer state.
func (r *Renderer) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 20, G: 20, B: 30, A: 255})

	// Background layer (with transition)
	if r.transActive && r.oldBG != nil {
		r.drawImageScaled(screen, r.oldBG, 1.0-r.transProgress)
		r.drawImageScaled(screen, r.bgImage, r.transProgress)
	} else if r.bgImage != nil {
		r.drawImageScaled(screen, r.bgImage, r.bgAlpha)
	}

	// Character layer
	for _, cs := range r.chars {
		if cs.Image == nil {
			continue
		}
		r.drawCharacter(screen, cs)
	}

	// Foreground overlay (reserved)
	if r.fgOverlay != nil && r.fgAlpha > 0 {
		r.drawImageScaled(screen, r.fgOverlay, r.fgAlpha)
	}
}

// drawImageScaled draws an image scaled to fill the virtual resolution.
func (r *Renderer) drawImageScaled(dst *ebiten.Image, src *ebiten.Image, alpha float64) {
	if src == nil || alpha <= 0 {
		return
	}
	op := &ebiten.DrawImageOptions{}
	bw := float64(src.Bounds().Dx())
	bh := float64(src.Bounds().Dy())
	op.GeoM.Scale(VirtualWidth/bw, VirtualHeight/bh)
	op.ColorScale.ScaleAlpha(float32(alpha))
	dst.DrawImage(src, op)
}

// drawCharacter draws a character sprite at its preset position.
func (r *Renderer) drawCharacter(dst *ebiten.Image, cs *CharSprite) {
	op := &ebiten.DrawImageOptions{}
	iw := float64(cs.Image.Bounds().Dx())
	ih := float64(cs.Image.Bounds().Dy())

	// Scale character to roughly 60% of screen height, preserving aspect ratio
	targetH := VirtualHeight * 0.6
	scale := targetH / ih
	scaledW := iw * scale

	op.GeoM.Scale(scale, scale)

	// Position horizontally
	var x float64
	switch cs.Position {
	case PosLeft:
		x = VirtualWidth*0.15 - scaledW/2
	case PosRight:
		x = VirtualWidth*0.85 - scaledW/2
	default: // center
		x = (VirtualWidth - scaledW) / 2
	}
	// Anchor to bottom of screen
	y := VirtualHeight - targetH

	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleAlpha(float32(cs.Alpha))
	dst.DrawImage(cs.Image, op)
}
