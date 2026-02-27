package render

import (
	"image/color"
	"regexp"
	"strconv"
	"strings"
)

// TextSegment represents a styled piece of text.
type TextSegment struct {
	Text  string
	Color color.Color
	Size  float64 // 0.0 means default
	Pause bool    // [w] marker — pause typewriter here
}

var (
	colorTagRe  = regexp.MustCompile(`<color=([^>]+)>`)
	closeTagRe  = regexp.MustCompile(`</color>`)
	sizeTagRe   = regexp.MustCompile(`<size=([0-9]+)>`)
	closeSizeRe = regexp.MustCompile(`</size>`)
	pauseRe     = regexp.MustCompile(`\[w\]`)
)

// colorMap maps color names to RGBA values.
var colorMap = map[string]color.Color{
	"red":    color.RGBA{R: 255, G: 80, B: 80, A: 255},
	"blue":   color.RGBA{R: 100, G: 150, B: 255, A: 255},
	"green":  color.RGBA{R: 100, G: 255, B: 100, A: 255},
	"yellow": color.RGBA{R: 255, G: 255, B: 100, A: 255},
	"white":  color.White,
	"gray":   color.RGBA{R: 180, G: 180, B: 180, A: 255},
	"pink":   color.RGBA{R: 255, G: 150, B: 200, A: 255},
	"cyan":   color.RGBA{R: 100, G: 255, B: 255, A: 255},
	"orange": color.RGBA{R: 255, G: 180, B: 50, A: 255},
}

// ParseRichText parses a string with <color=X>...</color> and [w] markers
// into a list of styled segments.
func ParseRichText(s string) []TextSegment {
	var segments []TextSegment
	currentColor := color.Color(color.White)
	currentSize := 0.0 // 0 means default
	remaining := s

	for len(remaining) > 0 {
		// Find the earliest tag
		colorLoc := colorTagRe.FindStringIndex(remaining)
		closeLoc := closeTagRe.FindStringIndex(remaining)
		sizeLoc := sizeTagRe.FindStringIndex(remaining)
		closeSizeLoc := closeSizeRe.FindStringIndex(remaining)
		pauseLoc := pauseRe.FindStringIndex(remaining)

		earliest := len(remaining)
		tagType := ""

		if colorLoc != nil && colorLoc[0] < earliest {
			earliest = colorLoc[0]
			tagType = "color"
		}
		if closeLoc != nil && closeLoc[0] < earliest {
			earliest = closeLoc[0]
			tagType = "close"
		}
		if sizeLoc != nil && sizeLoc[0] < earliest {
			earliest = sizeLoc[0]
			tagType = "size"
		}
		if closeSizeLoc != nil && closeSizeLoc[0] < earliest {
			earliest = closeSizeLoc[0]
			tagType = "closeSize"
		}
		if pauseLoc != nil && pauseLoc[0] < earliest {
			earliest = pauseLoc[0]
			tagType = "pause"
		}

		// Add text before the tag
		if earliest > 0 {
			segments = append(segments, TextSegment{
				Text:  remaining[:earliest],
				Color: currentColor,
				Size:  currentSize,
			})
		}

		switch tagType {
		case "color":
			match := colorTagRe.FindStringSubmatch(remaining[colorLoc[0]:])
			if len(match) > 1 {
				name := strings.ToLower(match[1])
				if c, ok := colorMap[name]; ok {
					currentColor = c
				}
			}
			remaining = remaining[colorLoc[1]:]
		case "close":
			currentColor = color.White
			remaining = remaining[closeLoc[1]:]
		case "size":
			match := sizeTagRe.FindStringSubmatch(remaining[sizeLoc[0]:])
			if len(match) > 1 {
				if s, err := strconv.ParseFloat(match[1], 64); err == nil {
					currentSize = s
				}
			}
			remaining = remaining[sizeLoc[1]:]
		case "closeSize":
			currentSize = 0
			remaining = remaining[closeSizeLoc[1]:]
		case "pause":
			segments = append(segments, TextSegment{Pause: true})
			remaining = remaining[pauseLoc[1]:]
		default:
			// No more tags
			remaining = ""
		}
	}

	return segments
}

// StripTags removes all rich text tags, returning plain text and pause positions (rune indices).
func StripTags(s string) (plain string, pauses []int) {
	segments := ParseRichText(s)
	var b strings.Builder
	runeCount := 0
	for _, seg := range segments {
		if seg.Pause {
			pauses = append(pauses, runeCount)
			continue
		}
		b.WriteString(seg.Text)
		runeCount += len([]rune(seg.Text))
	}
	return b.String(), pauses
}
