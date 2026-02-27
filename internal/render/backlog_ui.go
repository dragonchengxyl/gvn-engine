package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// BacklogEntry mirrors the structure needed for display.
// We pass this data in when showing the backlog.
type BacklogEntry struct {
	Speaker string
	Content string
}

// BacklogUI handles the history log display.
type BacklogUI struct {
	faceSource *text.GoTextFaceSource
	visible    bool
	entries    []BacklogEntry
	scrollY    float64
	maxScroll  float64

	// Layout constants
	x, y, w, h float32
}

func NewBacklogUI(faceSource *text.GoTextFaceSource) *BacklogUI {
	return &BacklogUI{
		faceSource: faceSource,
		x:          100,
		y:          100,
		w:          VirtualWidth - 200,
		h:          VirtualHeight - 200,
	}
}

func (ui *BacklogUI) Show(entries []BacklogEntry) {
	ui.entries = entries
	ui.visible = true
	ui.scrollY = 0 // Reset scroll to top? Or bottom? usually bottom.

	// Calculate total height to set initial scroll to bottom
	totalH := ui.calculateTotalHeight()
	if totalH > float64(ui.h) {
		ui.scrollY = totalH - float64(ui.h)
	} else {
		ui.scrollY = 0
	}
}

func (ui *BacklogUI) Hide() {
	ui.visible = false
}

func (ui *BacklogUI) IsVisible() bool {
	return ui.visible
}

func (ui *BacklogUI) Scroll(delta float64) {
	if !ui.visible {
		return
	}
	ui.scrollY -= delta * 40 // Scroll speed
	ui.clampScroll()
}

func (ui *BacklogUI) clampScroll() {
	totalH := ui.calculateTotalHeight()
	maxS := totalH - float64(ui.h)
	if maxS < 0 {
		maxS = 0
	}
	if ui.scrollY < 0 {
		ui.scrollY = 0
	}
	if ui.scrollY > maxS {
		ui.scrollY = maxS
	}
}

func (ui *BacklogUI) calculateTotalHeight() float64 {
	h := 20.0 // padding
	face := &text.GoTextFace{Source: ui.faceSource, Size: 24}

	for _, e := range ui.entries {
		// Speaker
		h += 30
		// Content (multiline)
		width := float64(ui.w) - 40
		_, textH := text.Measure(e.Content, face, 0) // rough estimate if single line
		// We need to wrap text to measure height accurately
		// For now, let's estimate or use the same wrapping logic as TextRenderer
		// Using a simple approximation for MVP:
		lines := float64(len(e.Content)) * 24 / width // very rough
		if lines < 1 {
			lines = 1
		}
		h += textH*(lines+1) + 20 // spacing
	}
	return h
}

func (ui *BacklogUI) Draw(screen *ebiten.Image) {
	if !ui.visible {
		return
	}

	// Dim background
	vector.DrawFilledRect(screen, 0, 0, VirtualWidth, VirtualHeight, color.RGBA{0, 0, 0, 150}, false)

	// Panel background
	vector.DrawFilledRect(screen, ui.x, ui.y, ui.w, ui.h, color.RGBA{30, 30, 40, 240}, false)

	// Clip content to panel area?
	// Ebiten doesn't have easy clipping without sub-images.
	// For MVP, we just draw and let it overflow or handle strictly.
	// Let's try to render to a sub-image or just be careful.
	// Actually, `SubImage` is the way for clipping.

	// Define clipping area
	panelRect := ebiten.NewImage(int(ui.w), int(ui.h))
	// panelRect.Fill(color.RGBA{30, 30, 40, 255}) // Base color

	yOffset := -ui.scrollY + 20

	faceSpeaker := &text.GoTextFace{Source: ui.faceSource, Size: 28}
	faceContent := &text.GoTextFace{Source: ui.faceSource, Size: 24}

	for _, e := range ui.entries {
		// Draw Speaker
		if yOffset > -50 && yOffset < float64(ui.h) {
			op := &text.DrawOptions{}
			op.GeoM.Translate(20, yOffset)
			op.ColorScale.ScaleWithColor(color.RGBA{255, 200, 100, 255})
			text.Draw(panelRect, e.Speaker, faceSpeaker, op)
		}
		yOffset += 35

		// Draw Content
		// TODO: proper wrapping needed here for real long text
		// MVP: just draw it
		if yOffset > -100 && yOffset < float64(ui.h) {
			op := &text.DrawOptions{}
			op.GeoM.Translate(40, yOffset)
			op.ColorScale.ScaleWithColor(color.White)
			text.Draw(panelRect, e.Content, faceContent, op)
		}
		yOffset += 40 + 20 // line height + spacing
	}

	// Draw panel to screen
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(ui.x), float64(ui.y))
	screen.DrawImage(panelRect, op)

	// Scrollbar
	totalH := math.Max(ui.calculateTotalHeight(), float64(ui.h))
	scrollRatio := ui.scrollY / (totalH - float64(ui.h))
	if totalH > float64(ui.h) {
		barH := float32(ui.h * ui.h / float32(totalH))
		barY := ui.y + float32(scrollRatio)*(ui.h-barH)
		vector.DrawFilledRect(screen, ui.x+ui.w-10, barY, 8, barH, color.RGBA{100, 100, 100, 255}, false)
	}
}
