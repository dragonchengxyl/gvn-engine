package render

import (
	"bytes"
	"image/color"
	"io/fs"
	"log"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	// Text box layout (in virtual coordinates)
	textBoxX      = 80
	textBoxY      = 780
	textBoxW      = 1760
	textBoxH      = 260
	textBoxPad    = 24
	speakerFontSz = 28
	bodyFontSz    = 26
	lineSpacing   = 38
)

// TextRenderer handles text display with typewriter effect and CJK auto-wrap.
// It caches rendered text to a buffer image, only redrawing when content changes.
type TextRenderer struct {
	faceSource *text.GoTextFaceSource

	// Buffer for text caching (SKILL.md: only redraw when text changes)
	buffer      *ebiten.Image
	bufDirty    bool
	lastText    string
	lastLen     int
	lastSpeaker string
}

// NewTextRenderer creates a TextRenderer.
// Tries to load a TTF from the embedded FS; falls back to Go built-in font.
func NewTextRenderer(fsys fs.FS, fontPath string) *TextRenderer {
	tr := &TextRenderer{
		buffer: ebiten.NewImage(textBoxW, textBoxH),
	}

	// Try user-provided font first
	if data, err := fs.ReadFile(fsys, fontPath); err == nil {
		if src, err := text.NewGoTextFaceSource(bytes.NewReader(data)); err == nil {
			tr.faceSource = src
			log.Printf("[INFO] text: loaded font %q", fontPath)
			return tr
		}
		log.Printf("[WARN] text: failed to parse font %q — using built-in", fontPath)
	}

	// Fallback: Go built-in font (goregular — Latin only, CJK needs a real font)
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Printf("[ERROR] text: failed to load built-in font: %v", err)
		return tr
	}
	tr.faceSource = src
	log.Printf("[INFO] text: using built-in Go font (add a CJK .ttf to assets/fonts/ for Chinese)")
	return tr
}

// FaceSource returns the font source for sharing with other renderers.
func (tr *TextRenderer) FaceSource() *text.GoTextFaceSource {
	return tr.faceSource
}

// Draw renders the text box onto the screen.
func (tr *TextRenderer) Draw(screen *ebiten.Image, speaker, fullText string, visibleLen int) {
	if fullText == "" {
		return
	}

	// Semi-transparent text box background
	vector.DrawFilledRect(screen,
		float32(textBoxX), float32(textBoxY),
		float32(textBoxW), float32(textBoxH),
		color.RGBA{R: 0, G: 0, B: 0, A: 180}, false)

	// Strip tags to get plain text, then clamp visible length
	plain, _ := StripTags(fullText)
	runes := []rune(plain)
	if visibleLen > len(runes) {
		visibleLen = len(runes)
	}

	// Build the displayed rich text by truncating to visibleLen plain runes
	displayed := truncateRichText(fullText, visibleLen)

	// Check if buffer needs redraw
	if displayed != tr.lastText || speaker != tr.lastSpeaker || visibleLen != tr.lastLen {
		tr.bufDirty = true
		tr.lastText = displayed
		tr.lastSpeaker = speaker
		tr.lastLen = visibleLen
	}

	if tr.bufDirty {
		tr.redrawBuffer(speaker, displayed)
		tr.bufDirty = false
	}

	// Blit buffer onto screen
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(textBoxX, textBoxY)
	screen.DrawImage(tr.buffer, op)
}

// truncateRichText returns the rich text string truncated to n plain-text runes,
// preserving color tags for the visible portion.
func truncateRichText(s string, n int) string {
	segments := ParseRichText(s)
	var result string
	remaining := n
	for _, seg := range segments {
		if seg.Pause {
			continue
		}
		runes := []rune(seg.Text)
		if remaining <= 0 {
			break
		}
		if len(runes) <= remaining {
			result += seg.Text
			remaining -= len(runes)
		} else {
			result += string(runes[:remaining])
			remaining = 0
		}
	}
	return result
}

// redrawBuffer re-renders text content into the cached buffer image.
func (tr *TextRenderer) redrawBuffer(speaker, displayed string) {
	tr.buffer.Clear()

	if tr.faceSource == nil {
		return
	}

	yOffset := float64(textBoxPad)

	// Speaker name
	if speaker != "" {
		face := &text.GoTextFace{
			Source: tr.faceSource,
			Size:   speakerFontSz,
		}
		drawOp := &text.DrawOptions{}
		drawOp.GeoM.Translate(float64(textBoxPad), yOffset)
		drawOp.ColorScale.ScaleWithColor(color.RGBA{R: 255, G: 220, B: 100, A: 255})
		text.Draw(tr.buffer, speaker, face, drawOp)
		yOffset += lineSpacing + 4
	}

	// Body text with rich text color support
	face := &text.GoTextFace{
		Source: tr.faceSource,
		Size:   bodyFontSz,
	}
	maxW := float64(textBoxW - textBoxPad*2)

	// Parse rich text segments
	segments := ParseRichText(displayed)

	// Flatten segments into a single plain string for wrapping
	var plain string
	for _, seg := range segments {
		if !seg.Pause {
			plain += seg.Text
		}
	}

	lines := wrapText(plain, face, maxW)

	// Build a color map: rune index -> color
	colorIndex := make(map[int]color.Color)
	sizeIndex := make(map[int]float64)
	runePos := 0
	for _, seg := range segments {
		if seg.Pause {
			continue
		}
		for range []rune(seg.Text) {
			colorIndex[runePos] = seg.Color
			if seg.Size > 0 {
				sizeIndex[runePos] = seg.Size
			}
			runePos++
		}
	}

	// Render lines with per-character color
	globalRuneIdx := 0
	for _, line := range lines {
		xOffset := float64(textBoxPad)
		for _, r := range line {
			c := color.Color(color.White)
			if cc, ok := colorIndex[globalRuneIdx]; ok {
				c = cc
			}

			currentFace := face
			yDiff := 0.0
			if sz, ok := sizeIndex[globalRuneIdx]; ok {
				currentFace = &text.GoTextFace{
					Source: tr.faceSource,
					Size:   sz,
				}
				// Simple baseline adjustment approximation
				yDiff = (bodyFontSz - sz) * 0.8
			}

			ch := string(r)
			drawOp := &text.DrawOptions{}
			drawOp.GeoM.Translate(xOffset, yOffset+yDiff)
			drawOp.ColorScale.ScaleWithColor(c)
			text.Draw(tr.buffer, ch, currentFace, drawOp)
			xOffset += text.Advance(ch, currentFace)
			globalRuneIdx++
		}
		yOffset += lineSpacing
	}
}

// wrapText breaks text into lines that fit within maxWidth pixels.
// Handles CJK characters (break anywhere) and Latin words (break at spaces).
func wrapText(s string, face *text.GoTextFace, maxWidth float64) []string {
	if s == "" {
		return nil
	}

	var lines []string
	var current []rune

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\n' {
			lines = append(lines, string(current))
			current = current[:0]
			continue
		}

		current = append(current, r)
		w := text.Advance(string(current), face)
		if w > maxWidth && len(current) > 1 {
			if isCJK(r) {
				lines = append(lines, string(current[:len(current)-1]))
				current = []rune{r}
			} else {
				idx := lastSpaceIndex(current)
				if idx > 0 {
					lines = append(lines, string(current[:idx]))
					current = current[idx+1:]
				} else {
					lines = append(lines, string(current[:len(current)-1]))
					current = []rune{r}
				}
			}
		}
	}
	if len(current) > 0 {
		lines = append(lines, string(current))
	}
	return lines
}

func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hangul, r) ||
		(r >= 0x3000 && r <= 0x303F)
}

func lastSpaceIndex(runes []rune) int {
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] == ' ' {
			return i
		}
	}
	return -1
}

// ForceRefresh marks the text buffer as dirty so it is redrawn on the next Draw call.
// Call this after a save/load to ensure the restored text is re-rendered.
func (tr *TextRenderer) ForceRefresh() {
	tr.bufDirty = true
	tr.lastText = ""
	tr.lastSpeaker = ""
	tr.lastLen = -1
}
