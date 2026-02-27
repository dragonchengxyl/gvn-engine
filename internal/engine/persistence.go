package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gvn-engine/internal/render"
)

const saveDirName = "saves"

// SaveGame saves the current context to a JSON file in the saves directory.
func (g *Game) SaveGame(slot int) error {
	// Create saves directory if it doesn't exist
	if err := os.MkdirAll(saveDirName, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	data, err := json.MarshalIndent(g.Ctx, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	fname := filepath.Join(saveDirName, fmt.Sprintf("save_%d.json", slot))
	if err := os.WriteFile(fname, data, 0644); err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}
	return nil
}

// LoadGame loads the context from a save file and restores visual state.
func (g *Game) LoadGame(slot int) error {
	fname := filepath.Join(saveDirName, fmt.Sprintf("save_%d.json", slot))
	data, err := os.ReadFile(fname)
	if err != nil {
		return fmt.Errorf("failed to read save file: %w", err)
	}

	newCtx := NewContext()
	if err := json.Unmarshal(data, newCtx); err != nil {
		return fmt.Errorf("failed to unmarshal context: %w", err)
	}

	g.Ctx = newCtx

	// CRITICAL: Update references in UI components that point to Context fields
	g.Config.SetReferences(&g.Audio.MasterVolume, &g.Ctx.TextSpeed)

	g.restoreVisualState()
	return nil
}

// restoreVisualState resynchronizes the Renderer and Audio with the Context data.
func (g *Game) restoreVisualState() {
	// 1. Restore Background
	if g.Ctx.Background != "" {
		img := g.Assets.LoadImage("images/" + g.Ctx.Background)
		// Immediate set, no transition
		g.Renderer.SetBackground(img, 0)
	}

	// 2. Restore Characters
	g.Renderer.ClearCharacters()
	for name, state := range g.Ctx.Characters {
		img := g.Assets.LoadImage("images/" + state.File)
		g.Renderer.SetCharacter(name, img, render.CharPosition(state.Position))
	}

	// 3. Restore BGM
	if g.Ctx.CurrentBGM != "" {
		g.Audio.PlayBGM("audio/" + g.Ctx.CurrentBGM)
	}

	// 4. Reset Engine State
	// If we saved during typing, we might want to restore text instantly or restart typing.
	// For simplicity, let's just show the full text and wait.
	if g.Ctx.FullText != "" {
		g.State = StateWaitInput
		g.Ctx.VisibleLen = len(g.Ctx.FullText) // Show full text
	} else {
		g.State = StateIdle
	}

	// 5. Restore Foreground
	if g.Ctx.Foreground != "" {
		img := g.Assets.LoadImage("images/" + g.Ctx.Foreground)
		g.Renderer.SetForeground(img, g.Ctx.FGAlpha)
	} else {
		g.Renderer.ClearForeground()
	}

	// 6. Force Text Redraw
	g.Text.ForceRefresh()
}

// ScanSaveSlots checks the saves directory and updates the UI with file info.
func (g *Game) ScanSaveSlots() {
	for i := 1; i <= 12; i++ {
		fname := filepath.Join(saveDirName, fmt.Sprintf("save_%d.json", i))
		info, err := os.Stat(fname)
		if err != nil {
			g.SaveLoad.SetSlotInfo(i, "(Empty)")
			continue
		}
		// Format: YYYY-MM-DD HH:MM
		timeStr := info.ModTime().Format("2006-01-02 15:04")
		g.SaveLoad.SetSlotInfo(i, timeStr)
	}
}
