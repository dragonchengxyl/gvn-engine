package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	choiceBoxW   = 600
	choiceBoxH   = 60
	choiceGap    = 20
	choiceFontSz = 24
)

// Choice represents a single selectable option.
type Choice struct {
	Text   string
	Label  string // jump target
	SetKey string // variable key to set on selection
	SetVal string // variable value to set on selection
}

// ChoiceRenderer draws interactive choice buttons on screen.
type ChoiceRenderer struct {
	faceSource *text.GoTextFaceSource
	choices    []Choice
	active     bool
	rects      []choiceRect // computed hit areas
}

type choiceRect struct {
	X, Y, W, H int
}

// NewChoiceRenderer creates a choice renderer sharing the same font source.
func NewChoiceRenderer(faceSource *text.GoTextFaceSource) *ChoiceRenderer {
	return &ChoiceRenderer{
		faceSource: faceSource,
	}
}

// Show displays a set of choices on screen.
func (cr *ChoiceRenderer) Show(choices []Choice) {
	cr.choices = choices
	cr.active = true
	cr.computeRects()
}

// Hide clears the choice display.
func (cr *ChoiceRenderer) Hide() {
	cr.choices = nil
	cr.active = false
	cr.rects = nil
}

// IsActive returns true if choices are currently displayed.
func (cr *ChoiceRenderer) IsActive() bool {
	return cr.active
}

// HitTest checks if a click at (x, y) hits any choice button.
// Returns the index of the hit choice, or -1 if none.
func (cr *ChoiceRenderer) HitTest(x, y int) int {
	for i, r := range cr.rects {
		if x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H {
			return i
		}
	}
	return -1
}

// Draw renders the choice buttons onto the screen.
func (cr *ChoiceRenderer) Draw(screen *ebiten.Image) {
	if !cr.active || len(cr.choices) == 0 {
		return
	}

	for i, ch := range cr.choices {
		r := cr.rects[i]

		// Button background
		vector.DrawFilledRect(screen,
			float32(r.X), float32(r.Y),
			float32(r.W), float32(r.H),
			color.RGBA{R: 40, G: 40, B: 80, A: 220}, false)

		// Button border
		vector.StrokeRect(screen,
			float32(r.X), float32(r.Y),
			float32(r.W), float32(r.H),
			2, color.RGBA{R: 180, G: 180, B: 255, A: 255}, false)

		// Button text
		if cr.faceSource != nil {
			face := &text.GoTextFace{
				Source: cr.faceSource,
				Size:   choiceFontSz,
			}
			drawOp := &text.DrawOptions{}
			drawOp.GeoM.Translate(float64(r.X+20), float64(r.Y+18))
			drawOp.ColorScale.ScaleWithColor(color.White)
			text.Draw(screen, ch.Text, face, drawOp)
		}
	}
}

// computeRects calculates button positions centered on screen.
func (cr *ChoiceRenderer) computeRects() {
	cr.rects = make([]choiceRect, len(cr.choices))
	totalH := len(cr.choices)*choiceBoxH + (len(cr.choices)-1)*choiceGap
	startY := (VirtualHeight - totalH) / 2
	startX := (VirtualWidth - choiceBoxW) / 2

	for i := range cr.choices {
		cr.rects[i] = choiceRect{
			X: startX,
			Y: startY + i*(choiceBoxH+choiceGap),
			W: choiceBoxW,
			H: choiceBoxH,
		}
	}
}
