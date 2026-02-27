package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type SaveLoadMode int

const (
	ModeSave SaveLoadMode = iota
	ModeLoad
)

type SaveLoadUI struct {
	faceSource *text.GoTextFaceSource
	visible    bool
	Mode       SaveLoadMode
	x, y, w, h float32

	slots     []int // 1..10
	slotInfo  map[int]string
	hoverSlot int
}

func NewSaveLoadUI(faceSource *text.GoTextFaceSource) *SaveLoadUI {
	ui := &SaveLoadUI{
		faceSource: faceSource,
		x:          200,
		y:          100,
		w:          1520,
		h:          880,
		slotInfo:   make(map[int]string),
	}
	for i := 1; i <= 12; i++ {
		ui.slots = append(ui.slots, i)
		ui.slotInfo[i] = "(Empty)"
	}
	return ui
}

func (ui *SaveLoadUI) SetSlotInfo(slot int, info string) {
	ui.slotInfo[slot] = info
}

func (ui *SaveLoadUI) Show(mode SaveLoadMode) {
	ui.Mode = mode
	ui.visible = true
}

func (ui *SaveLoadUI) Hide() {
	ui.visible = false
}

func (ui *SaveLoadUI) IsVisible() bool {
	return ui.visible
}

// Update returns the slot number if clicked, otherwise 0.
func (ui *SaveLoadUI) Update(mouseX, mouseY int, justPressed bool) int {
	if !ui.visible {
		return 0
	}

	mx, my := float32(mouseX), float32(mouseY)
	ui.hoverSlot = 0

	// Grid layout: 3 columns, 4 rows
	cols := 3
	rows := 4
	cellW := (ui.w - 100) / float32(cols)
	cellH := (ui.h - 150) / float32(rows)
	startX := ui.x + 50
	startY := ui.y + 100

	clickedSlot := 0

	for i, slot := range ui.slots {
		col := i % cols
		row := i / cols

		bx := startX + float32(col)*cellW
		by := startY + float32(row)*cellH
		bw := cellW - 20
		bh := cellH - 20

		if mx >= bx && mx <= bx+bw && my >= by && my <= by+bh {
			ui.hoverSlot = slot
			if justPressed {
				clickedSlot = slot
			}
		}
	}

	// Close Button logic
	if justPressed {
		closeX, closeY := ui.x+ui.w-60, ui.y+20
		if mx >= closeX && mx <= closeX+40 && my >= closeY && my <= closeY+40 {
			ui.Hide()
		}
	}

	return clickedSlot
}

func (ui *SaveLoadUI) Draw(screen *ebiten.Image) {
	if !ui.visible {
		return
	}

	// Dim background
	vector.DrawFilledRect(screen, 0, 0, VirtualWidth, VirtualHeight, color.RGBA{0, 0, 0, 200}, false)

	// Panel
	vector.DrawFilledRect(screen, ui.x, ui.y, ui.w, ui.h, color.RGBA{50, 45, 60, 255}, false)

	face := &text.GoTextFace{Source: ui.faceSource, Size: 40}

	// Title
	title := "Save Game"
	if ui.Mode == ModeLoad {
		title = "Load Game"
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(ui.x+50), float64(ui.y+30))
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, title, face, op)

	// Close Button
	vector.DrawFilledRect(screen, ui.x+ui.w-60, ui.y+20, 40, 40, color.RGBA{200, 50, 50, 255}, false)

	// Slots
	cols := 3
	cellW := (ui.w - 100) / float32(cols)
	cellH := (ui.h - 150) / float32(4) // 4 rows
	startX := ui.x + 50
	startY := ui.y + 100

	smallFace := &text.GoTextFace{Source: ui.faceSource, Size: 24}

	for i, slot := range ui.slots {
		col := i % cols
		row := i / cols

		bx := startX + float32(col)*cellW
		by := startY + float32(row)*cellH
		bw := cellW - 20
		bh := cellH - 20

		bgColor := color.RGBA{70, 70, 80, 255}
		if ui.hoverSlot == slot {
			bgColor = color.RGBA{100, 100, 120, 255}
		}

		vector.DrawFilledRect(screen, bx, by, bw, bh, bgColor, false)

		// Slot Number
		op.GeoM.Reset()
		op.GeoM.Translate(float64(bx+10), float64(by+10))
		text.Draw(screen, fmt.Sprintf("Slot %d", slot), smallFace, op)

		// Info (Date/Time)
		info := ui.slotInfo[slot]
		op.GeoM.Reset()
		op.GeoM.Translate(float64(bx+10), float64(by+50))
		text.Draw(screen, info, smallFace, op)
	}
}
