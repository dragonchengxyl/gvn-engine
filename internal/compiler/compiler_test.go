package compiler

import (
	"testing"
)

// nvnSample is a complete .nvn script covering all syntax forms.
const nvnSample = `
// This is a comment — should be ignored.

# Opening

@bg: school_day.png --fade 1.5
@bgm: theme.ogg --loop true

[Sakura] (enter: left, expr: happy) "Hello! Welcome to GVN-Nexus."
[] "A nameless voice speaks from the void."
[Kyle] "I don't want to do this."

@se: beep.wav

>> What will you do?
  - Fight back -> Label_Fight
  - Turn and run -> Label_Flee

# Label_Fight
[Sakura] "Courage is the answer!"

# Label_Flee
[Sakura] "Sometimes retreat is wisdom."

@wait
`

func TestTokenize(t *testing.T) {
	tokens := Tokenize(nvnSample)

	var dialogs, systems, choices, options, labels, blanks int
	for _, tok := range tokens {
		switch tok.Kind {
		case TokDialog:
			dialogs++
		case TokSystem:
			systems++
		case TokChoice:
			choices++
		case TokOption:
			options++
		case TokLabel:
			labels++
		case TokBlank:
			blanks++
		}
	}

	if dialogs != 5 {
		t.Errorf("expected 5 dialog tokens, got %d", dialogs)
	}
	if systems != 4 {
		t.Errorf("expected 4 system tokens (@bg,@bgm,@se,@wait), got %d", systems)
	}
	if choices != 1 {
		t.Errorf("expected 1 choice token, got %d", choices)
	}
	if options != 2 {
		t.Errorf("expected 2 option tokens, got %d", options)
	}
	if labels != 3 {
		t.Errorf("expected 3 label tokens, got %d", labels)
	}
	if blanks == 0 {
		t.Error("expected at least 1 blank/comment token")
	}
}

func TestParseLabel(t *testing.T) {
	nodes := Parse("# Opening\n# Label_Fight")

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	for i, name := range []string{"Opening", "Label_Fight"} {
		n, ok := nodes[i].(*LabelNode)
		if !ok {
			t.Fatalf("node %d: expected *LabelNode", i)
		}
		if n.Name != name {
			t.Errorf("node %d: expected name %q, got %q", i, name, n.Name)
		}
	}
}

func TestParseSystemDirective(t *testing.T) {
	src := "@bg: school_day.png --fade 1.5\n@bgm: theme.ogg --loop true\n@wait"
	nodes := Parse(src)

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	bg := nodes[0].(*SystemNode)
	if bg.Command != "bg" {
		t.Errorf("expected command 'bg', got %q", bg.Command)
	}
	if bg.Value != "school_day.png" {
		t.Errorf("expected value 'school_day.png', got %q", bg.Value)
	}
	if bg.Options["fade"] != "1.5" {
		t.Errorf("expected fade=1.5, got %q", bg.Options["fade"])
	}

	bgm := nodes[1].(*SystemNode)
	if bgm.Options["loop"] != "true" {
		t.Errorf("expected loop=true, got %q", bgm.Options["loop"])
	}

	wait := nodes[2].(*SystemNode)
	if wait.Command != "wait" {
		t.Errorf("expected command 'wait', got %q", wait.Command)
	}
}

func TestParseDialog(t *testing.T) {
	tests := []struct {
		src           string
		wantSpeaker   string
		wantText      string
		wantParamKey  string
		wantParamVal  string
	}{
		{
			src:          `[Sakura] (enter: left, expr: happy) "Hello there!"`,
			wantSpeaker:  "Sakura",
			wantText:     "Hello there!",
			wantParamKey: "enter",
			wantParamVal: "left",
		},
		{
			src:         `[Kyle] "I don't want to do this."`,
			wantSpeaker: "Kyle",
			wantText:    "I don't want to do this.",
		},
		{
			src:         `[] "Anonymous line."`,
			wantSpeaker: "",
			wantText:    "Anonymous line.",
		},
	}

	for _, tt := range tests {
		nodes := Parse(tt.src)
		if len(nodes) != 1 {
			t.Fatalf("src %q: expected 1 node, got %d", tt.src, len(nodes))
		}
		d, ok := nodes[0].(*DialogNode)
		if !ok {
			t.Fatalf("src %q: expected *DialogNode", tt.src)
		}
		if d.Speaker != tt.wantSpeaker {
			t.Errorf("src %q: speaker: want %q, got %q", tt.src, tt.wantSpeaker, d.Speaker)
		}
		if d.Text != tt.wantText {
			t.Errorf("src %q: text: want %q, got %q", tt.src, tt.wantText, d.Text)
		}
		if tt.wantParamKey != "" {
			if d.Params[tt.wantParamKey] != tt.wantParamVal {
				t.Errorf("src %q: param[%s]: want %q, got %q",
					tt.src, tt.wantParamKey, tt.wantParamVal, d.Params[tt.wantParamKey])
			}
		}
	}
}

func TestParseChoice(t *testing.T) {
	src := `>> What will you do?
  - Fight back -> Label_Fight
  - Turn and run -> Label_Flee`

	nodes := Parse(src)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	c, ok := nodes[0].(*ChoiceNode)
	if !ok {
		t.Fatalf("expected *ChoiceNode, got %T", nodes[0])
	}
	if c.Prompt != "What will you do?" {
		t.Errorf("prompt: want %q, got %q", "What will you do?", c.Prompt)
	}
	if len(c.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(c.Options))
	}
	if c.Options[0].Text != "Fight back" {
		t.Errorf("option 0 text: want 'Fight back', got %q", c.Options[0].Text)
	}
	if c.Options[0].Label != "Label_Fight" {
		t.Errorf("option 0 label: want 'Label_Fight', got %q", c.Options[0].Label)
	}
	if c.Options[1].Label != "Label_Flee" {
		t.Errorf("option 1 label: want 'Label_Flee', got %q", c.Options[1].Label)
	}
}

func TestParseFullScript(t *testing.T) {
	nodes, err := Compile(nvnSample)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typeCount := map[NodeType]int{}
	for _, n := range nodes {
		typeCount[n.Type()]++
	}

	if typeCount[NodeDialog] != 5 {
		t.Errorf("expected 5 dialog nodes, got %d", typeCount[NodeDialog])
	}
	if typeCount[NodeSystem] != 4 {
		t.Errorf("expected 4 system nodes, got %d", typeCount[NodeSystem])
	}
	if typeCount[NodeChoice] != 1 {
		t.Errorf("expected 1 choice node, got %d", typeCount[NodeChoice])
	}
	if typeCount[NodeLabel] != 3 {
		t.Errorf("expected 3 label nodes, got %d", typeCount[NodeLabel])
	}
}

func TestCompileEmptySource(t *testing.T) {
	_, err := Compile("   \n  ")
	if err == nil {
		t.Error("expected error for empty source, got nil")
	}
}
