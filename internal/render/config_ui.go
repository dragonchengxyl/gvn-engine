package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type ConfigUI struct {
	faceSource *text.GoTextFaceSource
	visible    bool
	x, y, w, h float32

	// Settings references
	masterVolume *float64
	textSpeed    *int

	// UI State
	draggingVol bool
	draggingSpd bool
}

func NewConfigUI(faceSource *text.GoTextFaceSource, masterVolume *float64, textSpeed *int) *ConfigUI {
	return &ConfigUI{
		faceSource:   faceSource,
		x:            400,
		y:            200,
		w:            1120,
		h:            680,
		masterVolume: masterVolume,
		textSpeed:    textSpeed,
	}
}

func (ui *ConfigUI) Show() {
	ui.visible = true
}

func (ui *ConfigUI) SetReferences(masterVolume *float64, textSpeed *int) {
	ui.masterVolume = masterVolume
	ui.textSpeed = textSpeed
}

func (ui *ConfigUI) Hide() {
	ui.visible = false
}

func (ui *ConfigUI) IsVisible() bool {
	return ui.visible
}

func (ui *ConfigUI) Update(mouseX, mouseY int, justPressed bool, pressed bool) {
	if !ui.visible {
		return
	}

	mx, my := float32(mouseX), float32(mouseY)

	// Volume Slider Area
	volRectX, volRectY := ui.x+300, ui.y+150
	volRectW, volRectH := float32(400), float32(30)

	if pressed {
		if mx >= volRectX && mx <= volRectX+volRectW && my >= volRectY-10 && my <= volRectY+volRectH+10 {
			ui.draggingVol = true
		}
	} else {
		ui.draggingVol = false
	}

	if ui.draggingVol {
		ratio := float64((mx - volRectX) / volRectW)
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}
		*ui.masterVolume = ratio
	}

	// Text Speed Slider Area
	spdRectX, spdRectY := ui.x+300, ui.y+250

	if pressed {
		if mx >= spdRectX && mx <= spdRectX+volRectW && my >= spdRectY-10 && my <= spdRectY+volRectH+10 {
			ui.draggingSpd = true
		}
	} else {
		ui.draggingSpd = false
	}

	if ui.draggingSpd {
		ratio := float64((mx - spdRectX) / volRectW)
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}
		// Speed range: 10 (slow) to 100 (fast)
		*ui.textSpeed = int(10 + ratio*90)
	}

	// Close Button (HitTest logic simple here)
	if justPressed {
		closeX, closeY := ui.x+ui.w-60, ui.y+20
		if mx >= closeX && mx <= closeX+40 && my >= closeY && my <= closeY+40 {
			ui.Hide()
		}
	}
}

func (ui *ConfigUI) Draw(screen *ebiten.Image) {
	if !ui.visible {
		return
	}

	// Dim background
	vector.DrawFilledRect(screen, 0, 0, VirtualWidth, VirtualHeight, color.RGBA{0, 0, 0, 180}, false)

	// Panel
	vector.DrawFilledRect(screen, ui.x, ui.y, ui.w, ui.h, color.RGBA{40, 40, 50, 255}, false)

	face := &text.GoTextFace{Source: ui.faceSource, Size: 32}

	// Title
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(ui.x+40), float64(ui.y+40))
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, "Configuration", face, op)

	// Close Button [X]
	vector.DrawFilledRect(screen, ui.x+ui.w-60, ui.y+20, 40, 40, color.RGBA{200, 50, 50, 255}, false)

	// Volume Slider
	op.GeoM.Reset()
	op.GeoM.Translate(float64(ui.x+60), float64(ui.y+150))
	text.Draw(screen, "Master Volume", face, op)

	volRectX, volRectY := ui.x+300, ui.y+150
	volRectW, volRectH := float32(400), float32(30)
	vector.DrawFilledRect(screen, volRectX, volRectY, volRectW, volRectH, color.RGBA{80, 80, 80, 255}, false)

	// Knob
	volX := volRectX + float32(*ui.masterVolume)*volRectW
	vector.DrawFilledRect(screen, volX-10, volRectY-5, 20, 40, color.RGBA{100, 200, 100, 255}, false)

	op.GeoM.Reset()
	op.GeoM.Translate(float64(volRectX+volRectW+20), float64(volRectY))
	text.Draw(screen, fmt.Sprintf("%d%%", int(*ui.masterVolume*100)), face, op)

	// Text Speed Slider
	op.GeoM.Reset()
	op.GeoM.Translate(float64(ui.x+60), float64(ui.y+250))
	text.Draw(screen, "Text Speed", face, op)

	spdRectX, spdRectY := ui.x+300, ui.y+250
	vector.DrawFilledRect(screen, spdRectX, spdRectY, volRectW, volRectH, color.RGBA{80, 80, 80, 255}, false)

	// Knob
	spdRatio := float32(*ui.textSpeed-10) / 90.0
	spdX := spdRectX + spdRatio*volRectW
	vector.DrawFilledRect(screen, spdX-10, spdRectY-5, 20, 40, color.RGBA{100, 150, 250, 255}, false)

	op.GeoM.Reset()
	op.GeoM.Translate(float64(spdRectX+volRectW+20), float64(spdRectY))
	text.Draw(screen, fmt.Sprintf("%d CPS", *ui.textSpeed), face, op)
}
