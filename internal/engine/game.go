package engine

import (
	"fmt"
	"io/fs"
	"log"
	"strconv"

	gaudio "gvn-engine/internal/audio"
	"gvn-engine/internal/ecs"
	"gvn-engine/internal/history"
	"gvn-engine/internal/input"
	"gvn-engine/internal/loader"
	"gvn-engine/internal/render"
	"gvn-engine/internal/script"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Game implements ebiten.Game and drives the visual novel engine.
type Game struct {
	State    EngineState
	Ctx      *Context
	Script   *script.ScriptFile
	Assets   *loader.AssetManager
	Renderer *render.Renderer
	Text     *render.TextRenderer
	Input    *input.Handler
	Audio    *gaudio.Manager
	Choices  *render.ChoiceRenderer
	UI       *render.UIRenderer
	Backlog  *render.BacklogUI
	Config   *render.ConfigUI
	SaveLoad *render.SaveLoadUI

	// History stack — Event Sourcing (Phase 6)
	History *history.Stack

	// ECS World — Phase 7 animated entities
	ECS *ecs.World

	// Active choice data for StateChoice
	activeChoices []render.Choice

	// Tick counter for typewriter speed control
	typingTicks int
	// Tick counter for auto-mode delay
	waitTicks int
	// Pause positions from [w] markers (rune indices)
	pausePositions []int
	// Whether we're currently paused at a [w] marker
	pausedAtW   bool
	pauseWTicks int
}

// NewGame creates and initializes a Game instance.
func NewGame(sf *script.ScriptFile, assets *loader.AssetManager, assetsFS fs.FS) *Game {
	g := &Game{
		State:    StateIdle,
		Ctx:      NewContext(),
		Script:   sf,
		Assets:   assets,
		Renderer: render.NewRenderer(),
		Text:     render.NewTextRenderer(assetsFS, "fonts/default.ttf"),
		Input:    input.NewHandler(),
		Audio:    gaudio.NewManager(assetsFS),
	}
	g.Choices = render.NewChoiceRenderer(g.Text.FaceSource())
	g.UI = render.NewUIRenderer(g.Text.FaceSource())
	g.Backlog = render.NewBacklogUI(g.Text.FaceSource())
	g.Config = render.NewConfigUI(g.Text.FaceSource(), &g.Audio.MasterVolume, &g.Ctx.TextSpeed)
	g.SaveLoad = render.NewSaveLoadUI(g.Text.FaceSource())
	g.History = history.NewStack(g.Ctx.ToHistoryState())
	g.ECS = ecs.NewWorld()
	log.Printf("[INFO] engine: initialized with script %q (%d commands)", sf.Title, len(sf.Commands))
	return g
}

// Update runs at 60 TPS. Handles state transitions and command execution.
func (g *Game) Update() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[FATAL] engine: recovered from panic: %v", r)
			g.Ctx.Advance()
			g.State = StateIdle
		}
	}()

	// Update renderer transitions
	g.Renderer.Update()
	g.Input.Update()
	// Advance ECS tween animations (non-blocking, runs every tick)
	ecs.TweenSystem(g.ECS, 1.0/60.0)
	mouseX, mouseY := g.Input.Position()
	g.UI.Update(mouseX, mouseY)

	// Update Overlays
	justPressed := g.Input.JustPressed()
	pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) // simple check for dragging

	if g.Config.IsVisible() {
		g.Config.Update(mouseX, mouseY, justPressed, pressed)
		g.Audio.UpdateVolume() // Apply volume changes
		return nil             // Block other inputs
	}

	if g.SaveLoad.IsVisible() {
		slot := g.SaveLoad.Update(mouseX, mouseY, justPressed)
		if slot > 0 {
			log.Printf("[INFO] engine: SaveLoadUI clicked slot %d", slot)
			if g.SaveLoad.Mode == render.ModeSave {
				if err := g.SaveGame(slot); err != nil {
					log.Printf("[ERROR] engine: save failed: %v", err)
				} else {
					log.Printf("[INFO] engine: saved to slot %d", slot)
					g.SaveLoad.Hide()
				}
			} else {
				if err := g.LoadGame(slot); err != nil {
					log.Printf("[ERROR] engine: load failed: %v", err)
				} else {
					log.Printf("[INFO] engine: loaded from slot %d", slot)
					g.SaveLoad.Hide()
					// Force Text Redraw
					g.Text.ForceRefresh()
				}
			}
		}
		return nil
	}

	if g.Backlog.IsVisible() {
		_, dy := ebiten.Wheel()
		if dy != 0 {
			g.Backlog.Scroll(dy)
		}
		// Close on click outside? or just a button?
		// For now, click anywhere to close
		if justPressed {
			g.Backlog.Hide()
		}
		return nil
	}

	// Scroll-up → Undo: walk back one dialogue step (Time Machine)
	_, wheelY := ebiten.Wheel()
	if wheelY > 0 && g.History.CanUndo() &&
		(g.State == StateWaitInput || g.State == StateIdle) {
		g.applyHistoryUndo()
		return nil
	}

	// Handle UI clicks
	if g.Input.JustPressed() {
		action := g.UI.HitTest(mouseX, mouseY)
		switch action {
		case render.ActionAuto:
			g.Ctx.AutoMode = !g.Ctx.AutoMode
			if g.Ctx.AutoMode {
				g.Ctx.SkipMode = false
			}
			log.Printf("[INFO] engine: Auto Mode = %v", g.Ctx.AutoMode)
		case render.ActionSkip:
			g.Ctx.SkipMode = !g.Ctx.SkipMode
			if g.Ctx.SkipMode {
				g.Ctx.AutoMode = false
			}
			log.Printf("[INFO] engine: Skip Mode = %v", g.Ctx.SkipMode)
		case render.ActionSave:
			g.ScanSaveSlots()
			g.SaveLoad.Show(render.ModeSave)
		case render.ActionLoad:
			g.ScanSaveSlots()
			g.SaveLoad.Show(render.ModeLoad)
		case render.ActionLog:
			// Convert history to backlog entries
			var entries []render.BacklogEntry
			for _, h := range g.Ctx.History {
				entries = append(entries, render.BacklogEntry{Speaker: h.Speaker, Content: h.Content})
			}
			g.Backlog.Show(entries)
		case render.ActionConfig:
			g.Config.Show()
		}

		// Consume input if UI was clicked
		if action != render.ActionNone {
			return nil
		}
	}

	// Toggle Auto/Skip modes (Keyboard shortcuts)
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.Ctx.AutoMode = !g.Ctx.AutoMode
		if g.Ctx.AutoMode {
			g.Ctx.SkipMode = false
		}
		log.Printf("[INFO] engine: Auto Mode = %v", g.Ctx.AutoMode)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.Ctx.SkipMode = !g.Ctx.SkipMode
		if g.Ctx.SkipMode {
			g.Ctx.AutoMode = false
		}
		log.Printf("[INFO] engine: Skip Mode = %v", g.Ctx.SkipMode)
	}

	// Debug: Save/Load
	if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		if err := g.SaveGame(1); err != nil {
			log.Printf("[ERROR] engine: save failed: %v", err)
		} else {
			log.Printf("[INFO] engine: game saved to slot 1")
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
		if err := g.LoadGame(1); err != nil {
			log.Printf("[ERROR] engine: load failed: %v", err)
		} else {
			log.Printf("[INFO] engine: game loaded from slot 1")
		}
	}

	switch g.State {
	case StateIdle:
		g.executeNextCommand()
	case StateTyping:
		g.updateTyping()
	case StateWaitInput:
		if g.Input.JustPressed() {
			g.advanceLine()
			g.State = StateIdle
			g.waitTicks = 0
		} else if g.Ctx.SkipMode {
			// Skip: advance immediately
			g.advanceLine()
			g.State = StateIdle
			g.waitTicks = 0
		} else if g.Ctx.AutoMode {
			// Auto: advance after delay
			g.waitTicks++
			if g.waitTicks >= g.Ctx.AutoDelay {
				g.advanceLine()
				g.State = StateIdle
				g.waitTicks = 0
			}
		}
	case StateTransition:
		if !g.Renderer.IsTransitioning() {
			g.State = StateIdle
		}
	case StateChoice:
		if g.Input.JustPressed() {
			x, y := g.Input.Position()
			idx := g.Choices.HitTest(x, y)
			if idx >= 0 && idx < len(g.activeChoices) {
				chosen := g.activeChoices[idx]
				log.Printf("[INFO] engine: player chose %q -> %s", chosen.Text, chosen.Label)
				// Set variable if specified
				if chosen.SetKey != "" {
					g.Ctx.Variables[chosen.SetKey] = chosen.SetVal
					log.Printf("[INFO] engine: set %q = %q", chosen.SetKey, chosen.SetVal)
				}
				g.Choices.Hide()
				if chosen.Label != "" {
					g.jumpLine(chosen.Label)
				} else {
					g.advanceLine()
				}
				g.State = StateIdle
			}
		}
	}
	return nil
}

// Draw renders the current frame. Pure function of game state — no mutation.
func (g *Game) Draw(screen *ebiten.Image) {
	// Layer 1-3: BG, Characters, FG
	g.Renderer.Draw(screen)

	// Layer 3.5: ECS-driven animated entities (char fade-in, shader effects)
	render.DrawECSLayer(screen, g.ECS)

	// Layer 4: Text box
	g.Text.Draw(screen, g.Ctx.Speaker, g.Ctx.FullText, g.Ctx.VisibleLen)

	// Layer 5: Choice buttons
	g.Choices.Draw(screen)

	// Layer 6: UI HUD
	g.UI.Draw(screen)

	// Layer 7: Overlays (Backlog, Config, SaveLoad)
	if g.Backlog.IsVisible() {
		g.Backlog.Draw(screen)
	}
	if g.Config.IsVisible() {
		g.Config.Draw(screen)
	}
	if g.SaveLoad.IsVisible() {
		g.SaveLoad.Draw(screen)
	}

	// Debug HUD
	mode := ""
	if g.Ctx.AutoMode {
		mode = " | AUTO"
	} else if g.Ctx.SkipMode {
		mode = " | SKIP"
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("State: %s | Line: %d/%d | FPS: %.0f%s  [A]uto [S]kip",
		g.State, g.Ctx.LineIndex, len(g.Script.Commands), ebiten.ActualFPS(), mode), 4, 4)
}

// Layout returns the virtual resolution; Ebitengine handles scaling.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return render.VirtualWidth, render.VirtualHeight
}

// executeNextCommand processes the command at the current line index.
func (g *Game) executeNextCommand() {
	cmd := g.Ctx.CurrentCommand(g.Script)
	if cmd == nil {
		return
	}

	log.Printf("[INFO] engine: executing [%d] type=%s args=%v", g.Ctx.LineIndex, cmd.Type, cmd.Args)

	switch cmd.Type {
	case script.CmdBackground:
		file := cmd.Args["file"]
		if file != "" {
			g.applyAction(&history.ActionChangeBG{NewFile: file})
			img := g.Assets.LoadImage("images/" + file)
			g.Renderer.SetBackground(img, cmd.Duration)
		}
		g.advanceLine()
		if cmd.Duration > 0 {
			g.State = StateTransition
		}

	case script.CmdChar:
		name := cmd.Args["name"]
		file := cmd.Args["file"]
		pos := cmd.Args["position"]
		if name != "" && file != "" {
			g.applyAction(&history.ActionShowChar{Name: name, NewFile: file, NewPos: pos})
			img := g.Assets.LoadImage("images/" + file)
			g.Renderer.SetCharacter(name, img, render.CharPosition(pos))

			// Optional non-blocking fade-in via ECS Tween (--fade_in <seconds>)
			if fadeSec := parseFloat(cmd.Args["fade_in"]); fadeSec > 0 {
				g.spawnCharFade(name, img, fadeSec)
			}
		}
		g.advanceLine()

	case script.CmdText:
		speaker := cmd.Args["speaker"]
		rawContent := cmd.Args["content"]
		g.applyAction(&history.ActionShowText{NewSpeaker: speaker, NewText: rawContent})
		_, pauses := render.StripTags(rawContent)
		g.Ctx.VisibleLen = 0
		g.typingTicks = 0
		g.pausePositions = pauses
		g.pausedAtW = false
		g.pauseWTicks = 0
		g.State = StateTyping
		g.Ctx.History = append(g.Ctx.History, HistoryEntry{
			Speaker: g.Ctx.Speaker,
			Content: rawContent,
		})

	case script.CmdChoice:
		// Build choices from args (option_1/jump_1/set_key_1/set_val_1, ...)
		var choices []render.Choice
		for i := 1; i <= 4; i++ {
			key := fmt.Sprintf("option_%d", i)
			jumpKey := fmt.Sprintf("jump_%d", i)
			setKey := fmt.Sprintf("set_key_%d", i)
			setVal := fmt.Sprintf("set_val_%d", i)
			if text, ok := cmd.Args[key]; ok && text != "" {
				choices = append(choices, render.Choice{
					Text:   text,
					Label:  cmd.Args[jumpKey],
					SetKey: cmd.Args[setKey],
					SetVal: cmd.Args[setVal],
				})
			}
		}
		if len(choices) == 0 {
			log.Printf("[WARN] engine: choice command has no options, skipping")
			g.advanceLine()
		} else {
			g.activeChoices = choices
			g.Choices.Show(choices)
			g.State = StateChoice
		}

	case script.CmdSound:
		file := cmd.Args["file"]
		mode := cmd.Args["mode"] // "bgm" or "se" (default: "se")
		if mode == "bgm" {
			g.applyAction(&history.ActionPlayBGM{NewFile: file})
			g.Audio.PlayBGM("audio/" + file)
		} else {
			g.Audio.PlaySE("audio/" + file)
		}
		g.advanceLine()

	case script.CmdWait:
		g.State = StateWaitInput

	case script.CmdSet:
		k := cmd.Args["key"]
		v := cmd.Args["value"]
		if k != "" {
			g.applyAction(&history.ActionSetVar{Key: k, NewValue: v})
			log.Printf("[INFO] engine: set %q = %q", k, v)
		}
		g.advanceLine()

	case script.CmdIf:
		// Conditional jump: if args.key == args.value, jump to args.jump
		k := cmd.Args["key"]
		v := cmd.Args["value"]
		target := cmd.Args["jump"]
		elseTarget := cmd.Args["else"]
		if g.Ctx.Variables[k] == v {
			if target != "" {
				g.jumpLine(target)
			} else {
				g.advanceLine()
			}
		} else {
			if elseTarget != "" {
				g.jumpLine(elseTarget)
			} else {
				g.advanceLine()
			}
		}

	case script.CmdJump:
		// Unconditional jump to label
		target := cmd.Args["target"]
		if target != "" {
			g.jumpLine(target)
		} else {
			g.advanceLine()
		}

	case script.CmdFG:
		// Foreground overlay: file (image) + alpha (0.0~1.0), or "clear"
		file := cmd.Args["file"]
		if file == "clear" || file == "" {
			g.applyAction(&history.ActionSetFG{NewFile: "", NewAlpha: 0})
			g.Renderer.ClearForeground()
		} else {
			alpha := 1.0
			if aStr := cmd.Args["alpha"]; aStr != "" {
				if a, err := strconv.ParseFloat(aStr, 64); err == nil {
					alpha = a
				}
			}
			g.applyAction(&history.ActionSetFG{NewFile: file, NewAlpha: alpha})
			img := g.Assets.LoadImage("images/" + file)
			g.Renderer.SetForeground(img, alpha)
		}
		g.advanceLine()

	default:
		log.Printf("[WARN] engine: unhandled command type %q at line %d, skipping", cmd.Type, g.Ctx.LineIndex)
		g.advanceLine()
	}
}

// updateTyping advances the typewriter effect based on TextSpeed.
func (g *Game) updateTyping() {
	plain, _ := render.StripTags(g.Ctx.FullText)
	totalRunes := len([]rune(plain))

	// Skip mode: show all text instantly
	if g.Ctx.SkipMode {
		g.Ctx.VisibleLen = totalRunes
		g.State = StateWaitInput
		g.waitTicks = 0
		return
	}

	// Click to skip to end
	if g.Input.JustPressed() {
		g.Ctx.VisibleLen = totalRunes
		g.State = StateWaitInput
		g.waitTicks = 0
		return
	}

	// Handle [w] pause
	if g.pausedAtW {
		g.pauseWTicks++
		if g.pauseWTicks >= 30 { // 0.5 sec pause
			g.pausedAtW = false
			g.pauseWTicks = 0
		}
		return
	}

	// Advance at TextSpeed chars/sec (60 TPS)
	g.typingTicks++
	charsPerTick := float64(g.Ctx.TextSpeed) / 60.0
	newLen := int(float64(g.typingTicks) * charsPerTick)

	// Check for [w] pause at current position
	for _, p := range g.pausePositions {
		if g.Ctx.VisibleLen < p && newLen >= p {
			g.Ctx.VisibleLen = p
			g.pausedAtW = true
			g.pauseWTicks = 0
			return
		}
	}

	g.Ctx.VisibleLen = newLen
	if g.Ctx.VisibleLen >= totalRunes {
		g.Ctx.VisibleLen = totalRunes
		g.State = StateWaitInput
		g.waitTicks = 0
	}
}

// applyAction applies an action via the History stack and syncs Context data fields.
// Visual state (Renderer) is NOT updated here — callers handle that separately.
func (g *Game) applyAction(a history.Action) {
	if err := g.History.Apply(a); err != nil {
		log.Printf("[WARN] engine: history apply error: %v", err)
		return
	}
	g.Ctx.SyncFromHistoryState(g.History.State())
}

// applyHistoryUndo walks back one action, syncs Context + visual state.
func (g *Game) applyHistoryUndo() {
	if !g.History.Undo() {
		return
	}
	g.Ctx.SyncFromHistoryState(g.History.State())
	g.restoreVisualState()
	g.Text.ForceRefresh()
	g.State = StateWaitInput
	log.Printf("[INFO] engine: undo → line %d speaker=%q", g.Ctx.LineIndex, g.Ctx.Speaker)
}

// spawnCharFade creates an ECS entity that fades a character sprite from 0→1 alpha
// over fadeSec seconds. This is non-blocking: the engine does NOT enter StateTransition.
// The entity is automatically tagged "char_fade_<name>" and removed when the tween ends.
func (g *Game) spawnCharFade(name string, img *ebiten.Image, fadeSec float64) {
	tag := "char_fade_" + name
	if old := g.ECS.FindByTag(tag); old != nil {
		g.ECS.Remove(old.ID)
	}

	e := g.ECS.Create(tag)
	e.Sprite = &ecs.Sprite{Image: img}
	e.Alpha = &ecs.AlphaComp{Value: 0.0}

	alphaComp := e.Alpha
	id := e.ID
	e.Tween = ecs.EasedTween(0.0, 1.0, fadeSec, func(v float64) {
		alphaComp.Value = v
		if v >= 1.0 {
			g.ECS.Remove(id)
		}
	})
}

// parseFloat is a safe strconv.ParseFloat wrapper; returns 0 on error.
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// advanceLine records a line advance through the history system.
// Use this instead of g.Ctx.Advance() so history.state.LineIndex stays in sync.
func (g *Game) advanceLine() {
	g.applyAction(history.NewActionAdvanceLine(g.Ctx.LineIndex + 1))
}

// jumpLine records a label jump through the history system.
// Falls back to advanceLine if the label is not found.
func (g *Game) jumpLine(label string) {
	idx := g.Script.FindLabel(label)
	if idx < 0 {
		log.Printf("[WARN] engine: jump label %q not found, advancing instead", label)
		g.advanceLine()
		return
	}
	g.applyAction(history.NewActionAdvanceLine(idx))
}
