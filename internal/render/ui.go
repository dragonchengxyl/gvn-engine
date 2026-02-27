package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// UIAction defines the callback identifier for a button click.
type UIAction string

const (
	ActionNone     UIAction = ""
	ActionAuto     UIAction = "auto"
	ActionSkip     UIAction = "skip"
	ActionSave     UIAction = "save"
	ActionLoad     UIAction = "load"
	ActionLog      UIAction = "log"
	ActionConfig   UIAction = "config"
	ActionQuickSave UIAction = "qsave"
	ActionQuickLoad UIAction = "qload"
)

// Button represents a clickable UI element.
type Button struct {
	ID     UIAction
	Text   string
	X, Y   float32
	W, H   float32
	Color  color.Color
	Hover  bool
}

// UIRenderer manages HUD buttons and menus.
type UIRenderer struct {
	faceSource *text.GoTextFaceSource
	buttons    []*Button
	visible    bool
}

// NewUIRenderer creates a UI renderer.
func NewUIRenderer(faceSource *text.GoTextFaceSource) *UIRenderer {
	ui := &UIRenderer{
		faceSource: faceSource,
		visible:    true,
	}
	ui.initHUD()
	return ui
}

// initHUD creates the default in-game HUD buttons.
func (ui *UIRenderer) initHUD() {
	// Layout: Bottom right of the screen, above the text box? 
	// Or maybe a classic visual novel menu bar at the bottom right of the text box.
	// Let's put them in a row at the bottom right of the text box area.
	
	const btnW = 80
	const btnH = 30
	const startX = 1300
	const startY = 1040 // Bottom of screen (1080)
	const spacing = 90

	labels := []struct {
		Text string
		ID   UIAction
	}{
		{"Auto", ActionAuto},
		{"Skip", ActionSkip},
		{"Save", ActionSave},
		{"Load", ActionLoad},
		{"Log", ActionLog},
		{"Cfg", ActionConfig},
	}

	for i, l := range labels {
		ui.buttons = append(ui.buttons, &Button{
			ID:    l.ID,
			Text:  l.Text,
			X:     float32(startX + i*spacing),
			Y:     float32(startY),
			W:     btnW,
			H:     btnH,
			Color: color.RGBA{R: 50, G: 50, B: 60, A: 200},
		})
	}
}

// Update handles mouse hover states.
func (ui *UIRenderer) Update(mouseX, mouseY int) {
	if !ui.visible {
		return
	}
	mx, my := float32(mouseX), float32(mouseY)
	for _, btn := range ui.buttons {
		btn.Hover = mx >= btn.X && mx < btn.X+btn.W && my >= btn.Y && my < btn.Y+btn.H
	}
}

// HitTest checks if a button was clicked.
func (ui *UIRenderer) HitTest(x, y int) UIAction {
	if !ui.visible {
		return ActionNone
	}
	mx, my := float32(x), float32(y)
	for _, btn := range ui.buttons {
		if mx >= btn.X && mx < btn.X+btn.W && my >= btn.Y && my < btn.Y+btn.H {
			return btn.ID
		}
	}
	return ActionNone
}

// Draw renders the UI buttons.
func (ui *UIRenderer) Draw(screen *ebiten.Image) {
	if !ui.visible {
		return
	}

	for _, btn := range ui.buttons {
		// Background
		col := btn.Color
		if btn.Hover {
			col = color.RGBA{R: 100, G: 100, B: 120, A: 255}
		}
		vector.DrawFilledRect(screen, btn.X, btn.Y, btn.W, btn.H, col, false)

		// Text
		if ui.faceSource != nil {
			face := &text.GoTextFace{
				Source: ui.faceSource,
				Size:   18,
			}
			op := &text.DrawOptions{}
			op.GeoM.Translate(float64(btn.X+10), float64(btn.Y+6))
			op.ColorScale.ScaleWithColor(color.White)
			text.Draw(screen, btn.Text, face, op)
		}
	}
}
