// Package history implements Event Sourcing for the GVN engine.
// All state mutations go through Action.Apply; every Action carries an Undo.
// This package has zero dependency on ebiten or the engine package.
package history

// CharSnapshot stores a character's visual state.
type CharSnapshot struct {
	File     string
	Position string
}

// GameState is the minimal, fully-serialisable snapshot of game data
// that Actions operate on. It mirrors the relevant fields of engine.Context.
type GameState struct {
	Speaker    string
	FullText   string
	Background string
	Characters map[string]CharSnapshot // name -> snapshot
	Variables  map[string]string
	Foreground string
	FGAlpha    float64
	BGM        string
	LineIndex  int
}

// clone returns a deep copy of GameState (used inside actions to save "before" state).
func (s *GameState) clone() GameState {
	c := GameState{
		Speaker:    s.Speaker,
		FullText:   s.FullText,
		Background: s.Background,
		Foreground: s.Foreground,
		FGAlpha:    s.FGAlpha,
		BGM:        s.BGM,
		LineIndex:  s.LineIndex,
		Characters: make(map[string]CharSnapshot, len(s.Characters)),
		Variables:  make(map[string]string, len(s.Variables)),
	}
	for k, v := range s.Characters {
		c.Characters[k] = v
	}
	for k, v := range s.Variables {
		c.Variables[k] = v
	}
	return c
}

// Action is the core Event Sourcing interface.
// Every state change must be expressed as an Action.
type Action interface {
	Apply(state *GameState) error
	Undo(state *GameState) error
}

// ── Concrete Actions ────────────────────────────────────────────────────────

// ActionShowText advances the displayed text to a new speaker/content.
type ActionShowText struct {
	NewSpeaker string
	NewText    string
	// saved on first Apply
	prevSpeaker string
	prevText    string
	prevLine    int
}

func (a *ActionShowText) Apply(s *GameState) error {
	a.prevSpeaker = s.Speaker
	a.prevText = s.FullText
	a.prevLine = s.LineIndex
	s.Speaker = a.NewSpeaker
	s.FullText = a.NewText
	return nil
}

func (a *ActionShowText) Undo(s *GameState) error {
	s.Speaker = a.prevSpeaker
	s.FullText = a.prevText
	s.LineIndex = a.prevLine
	return nil
}

// ActionChangeBG transitions to a new background image.
type ActionChangeBG struct {
	NewFile string
	prevFile string
}

func (a *ActionChangeBG) Apply(s *GameState) error {
	a.prevFile = s.Background
	s.Background = a.NewFile
	return nil
}

func (a *ActionChangeBG) Undo(s *GameState) error {
	s.Background = a.prevFile
	return nil
}

// ActionShowChar adds or updates a character sprite.
type ActionShowChar struct {
	Name     string
	NewFile  string
	NewPos   string
	prevSnap *CharSnapshot // nil if character was absent before
}

func (a *ActionShowChar) Apply(s *GameState) error {
	if prev, ok := s.Characters[a.Name]; ok {
		snap := prev
		a.prevSnap = &snap
	} else {
		a.prevSnap = nil
	}
	s.Characters[a.Name] = CharSnapshot{File: a.NewFile, Position: a.NewPos}
	return nil
}

func (a *ActionShowChar) Undo(s *GameState) error {
	if a.prevSnap == nil {
		delete(s.Characters, a.Name)
	} else {
		s.Characters[a.Name] = *a.prevSnap
	}
	return nil
}

// ActionSetVar sets a script variable.
type ActionSetVar struct {
	Key      string
	NewValue string
	prevValue string
	prevExist bool
}

func (a *ActionSetVar) Apply(s *GameState) error {
	prev, ok := s.Variables[a.Key]
	a.prevValue = prev
	a.prevExist = ok
	s.Variables[a.Key] = a.NewValue
	return nil
}

func (a *ActionSetVar) Undo(s *GameState) error {
	if !a.prevExist {
		delete(s.Variables, a.Key)
	} else {
		s.Variables[a.Key] = a.prevValue
	}
	return nil
}

// ActionSetFG changes the foreground overlay.
type ActionSetFG struct {
	NewFile  string
	NewAlpha float64
	prevFile  string
	prevAlpha float64
}

func (a *ActionSetFG) Apply(s *GameState) error {
	a.prevFile = s.Foreground
	a.prevAlpha = s.FGAlpha
	s.Foreground = a.NewFile
	s.FGAlpha = a.NewAlpha
	return nil
}

func (a *ActionSetFG) Undo(s *GameState) error {
	s.Foreground = a.prevFile
	s.FGAlpha = a.prevAlpha
	return nil
}

// ActionPlayBGM changes the currently playing BGM track.
type ActionPlayBGM struct {
	NewFile  string
	prevFile string
}

func (a *ActionPlayBGM) Apply(s *GameState) error {
	a.prevFile = s.BGM
	s.BGM = a.NewFile
	return nil
}

func (a *ActionPlayBGM) Undo(s *GameState) error {
	s.BGM = a.prevFile
	return nil
}

// ActionAdvanceLine records a line-index advance (used for undo navigation).
type ActionAdvanceLine struct {
	prevIndex int
	newIndex  int
}

// NewActionAdvanceLine creates an ActionAdvanceLine with the given target index.
func NewActionAdvanceLine(newIndex int) *ActionAdvanceLine {
	return &ActionAdvanceLine{newIndex: newIndex}
}

func (a *ActionAdvanceLine) Apply(s *GameState) error {
	a.prevIndex = s.LineIndex
	s.LineIndex = a.newIndex
	return nil
}

func (a *ActionAdvanceLine) Undo(s *GameState) error {
	s.LineIndex = a.prevIndex
	return nil
}
