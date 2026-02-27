package engine

import (
	"log"

	"gvn-engine/internal/history"
	"gvn-engine/internal/script"
)

// CharState stores a character's visual state for serialization.
type CharState struct {
	File     string `json:"file"`
	Position string `json:"position"`
}

// HistoryEntry records a past dialogue line.
type HistoryEntry struct {
	Speaker string `json:"speaker"`
	Content string `json:"content"`
}

// Context holds all runtime data for the game.
// Designed to be fully serializable for save/load (json.Marshal).
type Context struct {
	// Script state
	ScriptFile string `json:"script_file"`
	LineIndex  int    `json:"line_index"`

	// Current visual state
	Background string               `json:"background"`
	Characters map[string]CharState `json:"characters"` // name -> state
	Foreground string               `json:"foreground"` // FG image file
	FGAlpha    float64              `json:"fg_alpha"`   // FG alpha

	// Game variables (Flag System)
	Variables map[string]string `json:"variables"`

	// Current text display
	Speaker    string         `json:"speaker"`
	FullText   string         `json:"full_text"`
	VisibleLen int            `json:"visible_len"` // for typewriter effect
	TextSpeed  int            `json:"text_speed"`  // chars per second
	History    []HistoryEntry `json:"history"`

	// Audio state
	CurrentBGM string `json:"current_bgm"`

	// Mode flags
	AutoMode  bool `json:"auto_mode"`
	SkipMode  bool `json:"skip_mode"`
	AutoDelay int  `json:"auto_delay"` // ticks to wait in auto mode (default 120 = 2 sec)
}

// NewContext creates a Context with sensible defaults.
func NewContext() *Context {
	return &Context{
		Characters: make(map[string]CharState),
		Variables:  make(map[string]string),
		History:    make([]HistoryEntry, 0),
		TextSpeed:  30,
		AutoDelay:  120, // 2 seconds at 60 TPS
	}
}

// Advance moves to the next command index.
func (c *Context) Advance() {
	c.LineIndex++
}

// CurrentCommand returns the command at the current line index, or nil if out of bounds.
func (c *Context) CurrentCommand(sf *script.ScriptFile) *script.ScriptCommand {
	if c.LineIndex < 0 || c.LineIndex >= len(sf.Commands) {
		return nil
	}
	return &sf.Commands[c.LineIndex]
}

// JumpTo sets the line index to the command with the given label.
// Returns false if the label is not found (logs a warning and stays put).
func (c *Context) JumpTo(sf *script.ScriptFile, label string) bool {
	idx := sf.FindLabel(label)
	if idx < 0 {
		log.Printf("[WARN] context: label %q not found, staying at line %d", label, c.LineIndex)
		return false
	}
	c.LineIndex = idx
	return true
}

// ToHistoryState builds a history.GameState snapshot from this Context.
func (c *Context) ToHistoryState() history.GameState {
	chars := make(map[string]history.CharSnapshot, len(c.Characters))
	for k, v := range c.Characters {
		chars[k] = history.CharSnapshot{File: v.File, Position: v.Position}
	}
	vars := make(map[string]string, len(c.Variables))
	for k, v := range c.Variables {
		vars[k] = v
	}
	return history.GameState{
		Speaker:    c.Speaker,
		FullText:   c.FullText,
		Background: c.Background,
		Characters: chars,
		Variables:  vars,
		Foreground: c.Foreground,
		FGAlpha:    c.FGAlpha,
		BGM:        c.CurrentBGM,
		LineIndex:  c.LineIndex,
	}
}

// SyncFromHistoryState copies data fields from a history.GameState back into this Context.
// Does NOT touch TextSpeed, AutoMode, SkipMode, History log, or VisibleLen.
func (c *Context) SyncFromHistoryState(hs *history.GameState) {
	c.Speaker = hs.Speaker
	c.FullText = hs.FullText
	c.Background = hs.Background
	c.Foreground = hs.Foreground
	c.FGAlpha = hs.FGAlpha
	c.CurrentBGM = hs.BGM
	c.LineIndex = hs.LineIndex
	c.Characters = make(map[string]CharState, len(hs.Characters))
	for k, v := range hs.Characters {
		c.Characters[k] = CharState{File: v.File, Position: v.Position}
	}
	c.Variables = make(map[string]string, len(hs.Variables))
	for k, v := range hs.Variables {
		c.Variables[k] = v
	}
}
