package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Handler provides unified mouse/touch input with proper just-pressed detection.
// Call Update() once per tick before checking state.
type Handler struct {
	// Last click/tap position in virtual coordinates
	X, Y int
	// Whether a click/tap just happened this frame
	justPressed bool
}

// NewHandler creates an input handler.
func NewHandler() *Handler {
	return &Handler{}
}

// Update polls input state. Must be called once per tick (in Game.Update).
func (h *Handler) Update() {
	h.justPressed = false

	// Mouse: just pressed this frame
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		h.X, h.Y = ebiten.CursorPosition()
		h.justPressed = true
		return
	}

	// Touch: any finger just pressed
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		h.X, h.Y = ebiten.TouchPosition(id)
		h.justPressed = true
		return
	}
}

// JustPressed returns true if the user clicked/tapped this frame.
func (h *Handler) JustPressed() bool {
	return h.justPressed
}

// Position returns the last click/tap position.
func (h *Handler) Position() (int, int) {
	return h.X, h.Y
}

// InRect returns true if the last click/tap was inside the given rectangle.
func (h *Handler) InRect(x, y, w, hh int) bool {
	if !h.justPressed {
		return false
	}
	return h.X >= x && h.X < x+w && h.Y >= y && h.Y < y+hh
}
