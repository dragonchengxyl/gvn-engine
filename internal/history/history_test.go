package history

import "testing"

func newTestState() GameState {
	return GameState{
		Speaker:    "",
		FullText:   "",
		Background: "",
		Characters: make(map[string]CharSnapshot),
		Variables:  make(map[string]string),
	}
}

func TestActionShowText(t *testing.T) {
	st := NewStack(newTestState())

	a := &ActionShowText{NewSpeaker: "Sakura", NewText: "Hello!"}
	if err := st.Apply(a); err != nil {
		t.Fatal(err)
	}
	if st.State().Speaker != "Sakura" || st.State().FullText != "Hello!" {
		t.Error("Apply did not set speaker/text")
	}

	if !st.Undo() {
		t.Error("Undo returned false")
	}
	if st.State().Speaker != "" || st.State().FullText != "" {
		t.Errorf("Undo did not restore: got speaker=%q text=%q", st.State().Speaker, st.State().FullText)
	}
}

func TestActionChangeBG(t *testing.T) {
	s := newTestState()
	s.Background = "old_bg.png"
	st := NewStack(s)

	if err := st.Apply(&ActionChangeBG{NewFile: "new_bg.png"}); err != nil {
		t.Fatal(err)
	}
	if st.State().Background != "new_bg.png" {
		t.Error("Apply did not change background")
	}
	st.Undo()
	if st.State().Background != "old_bg.png" {
		t.Errorf("Undo: expected 'old_bg.png', got %q", st.State().Background)
	}
}

func TestActionShowChar_NewAndUndo(t *testing.T) {
	st := NewStack(newTestState())

	if err := st.Apply(&ActionShowChar{Name: "Sakura", NewFile: "sakura.png", NewPos: "left"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := st.State().Characters["Sakura"]; !ok {
		t.Error("character not added after Apply")
	}

	st.Undo()
	if _, ok := st.State().Characters["Sakura"]; ok {
		t.Error("character should be removed after Undo (was absent before)")
	}
}

func TestActionShowChar_ReplaceAndUndo(t *testing.T) {
	s := newTestState()
	s.Characters["Sakura"] = CharSnapshot{File: "sakura_normal.png", Position: "center"}
	st := NewStack(s)

	st.Apply(&ActionShowChar{Name: "Sakura", NewFile: "sakura_sad.png", NewPos: "left"})
	if st.State().Characters["Sakura"].File != "sakura_sad.png" {
		t.Error("Apply did not update character")
	}

	st.Undo()
	if got := st.State().Characters["Sakura"].File; got != "sakura_normal.png" {
		t.Errorf("Undo: expected 'sakura_normal.png', got %q", got)
	}
}

func TestActionSetVar(t *testing.T) {
	st := NewStack(newTestState())

	st.Apply(&ActionSetVar{Key: "path", NewValue: "brave"})
	if st.State().Variables["path"] != "brave" {
		t.Error("variable not set")
	}

	st.Undo()
	if _, ok := st.State().Variables["path"]; ok {
		t.Error("variable should be absent after Undo (was not set before)")
	}
}

func TestActionSetFG(t *testing.T) {
	s := newTestState()
	s.Foreground = "rain.png"
	s.FGAlpha = 0.5
	st := NewStack(s)

	st.Apply(&ActionSetFG{NewFile: "", NewAlpha: 0})
	if st.State().Foreground != "" {
		t.Error("FG should be cleared after Apply")
	}

	st.Undo()
	if st.State().Foreground != "rain.png" || st.State().FGAlpha != 0.5 {
		t.Errorf("Undo did not restore FG: %q %.1f", st.State().Foreground, st.State().FGAlpha)
	}
}

func TestStackMultipleUndos(t *testing.T) {
	st := NewStack(newTestState())

	st.Apply(&ActionShowText{NewSpeaker: "A", NewText: "Line 1"})
	st.Apply(&ActionShowText{NewSpeaker: "B", NewText: "Line 2"})
	st.Apply(&ActionShowText{NewSpeaker: "C", NewText: "Line 3"})

	if st.Len() != 3 {
		t.Errorf("expected Len()=3, got %d", st.Len())
	}

	st.Undo()
	if st.State().Speaker != "B" {
		t.Errorf("after 1 undo: expected speaker B, got %q", st.State().Speaker)
	}
	st.Undo()
	if st.State().Speaker != "A" {
		t.Errorf("after 2 undos: expected speaker A, got %q", st.State().Speaker)
	}
	st.Undo()
	if st.State().Speaker != "" {
		t.Errorf("after 3 undos: expected empty speaker, got %q", st.State().Speaker)
	}

	// Extra Undo at boundary should return false
	if st.Undo() {
		t.Error("Undo past boundary should return false")
	}
	if st.CanUndo() {
		t.Error("CanUndo should be false at boundary")
	}
}

func TestStackApplyAfterUndo_DiscardsFuture(t *testing.T) {
	st := NewStack(newTestState())
	st.Apply(&ActionShowText{NewSpeaker: "A", NewText: "Line 1"})
	st.Apply(&ActionShowText{NewSpeaker: "B", NewText: "Line 2"})

	st.Undo() // back to A

	// Apply a new action — Line 2 branch should be discarded
	st.Apply(&ActionShowText{NewSpeaker: "C", NewText: "New branch"})

	if st.Len() != 2 {
		t.Errorf("expected Len()=2 after branch, got %d", st.Len())
	}
	if st.State().Speaker != "C" {
		t.Errorf("expected speaker C, got %q", st.State().Speaker)
	}
}
