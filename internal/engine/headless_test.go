package engine

import (
	"strings"
	"testing"
)

// sampleJSON is a minimal script covering all command types.
const sampleJSON = `{
  "title": "Headless Test",
  "commands": [
    {"type":"bg",     "args":{"file":"park.png"}, "duration":0.5},
    {"type":"char",   "args":{"name":"Alice","file":"alice.png","position":"left"}},
    {"type":"text",   "args":{"speaker":"Alice","content":"Hello, world!"}},
    {"type":"text",   "args":{"speaker":"","content":"Narrator line."}},
    {"type":"sound",  "args":{"file":"theme.ogg","mode":"bgm"}},
    {"type":"set",    "args":{"key":"flag","value":"on"}},
    {"type":"choice", "args":{"option_1":"Yes","jump_1":"end","option_2":"No","jump_2":"end"}},
    {"type":"text",   "args":{"speaker":"Alice","content":"You chose."}, "next":"end"},
    {"type":"jump",   "args":{"target":"end"}}
  ]
}`

const sampleNVN = `
# Opening

@bg: park.png --fade 0.5
@bgm: theme.ogg --loop true

[Alice] (enter: left) "Hello, world!"
[] "Narrator line."

>> What do you choose?
  - Yes -> Label_End
  - No  -> Label_End

# Label_End
[Alice] "You chose."
`

func TestRunHeadlessJSON(t *testing.T) {
	res, err := RunHeadlessJSON("test.json", []byte(sampleJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.DialogCount != 3 {
		t.Errorf("dialogs: want 3, got %d", res.DialogCount)
	}
	if res.ChoiceCount != 1 {
		t.Errorf("choices: want 1, got %d", res.ChoiceCount)
	}
	if res.TotalNodes != 9 {
		t.Errorf("total nodes: want 9, got %d", res.TotalNodes)
	}
	if len(res.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", res.Warnings)
	}
}

func TestRunHeadlessNVN(t *testing.T) {
	res, err := RunHeadlessNVN("test.nvn", sampleNVN)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.DialogCount != 3 {
		t.Errorf("dialogs: want 3, got %d", res.DialogCount)
	}
	if res.ChoiceCount != 1 {
		t.Errorf("choices: want 1, got %d", res.ChoiceCount)
	}
	if res.LabelCount != 2 {
		t.Errorf("labels: want 2, got %d", res.LabelCount)
	}
	if res.SystemCount != 2 { // @bg + @bgm
		t.Errorf("system nodes: want 2, got %d", res.SystemCount)
	}
}

func TestRunHeadlessAuto_JSON(t *testing.T) {
	res, err := RunHeadlessAuto("demo.json", []byte(sampleJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DialogCount == 0 {
		t.Error("expected at least 1 dialog from auto-dispatch JSON")
	}
}

func TestRunHeadlessAuto_NVN(t *testing.T) {
	res, err := RunHeadlessAuto("demo.nvn", []byte(sampleNVN))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DialogCount == 0 {
		t.Error("expected at least 1 dialog from auto-dispatch NVN")
	}
}

func TestRunHeadlessJSON_MissingLabel(t *testing.T) {
	badJSON := `{
  "title":"Bad",
  "commands":[
    {"type":"jump","args":{"target":"nonexistent"}}
  ]
}`
	res, err := RunHeadlessJSON("bad.json", []byte(badJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, w := range res.Warnings {
		if strings.Contains(w, "nonexistent") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning about missing label 'nonexistent', got: %v", res.Warnings)
	}
}

func TestHeadlessResultString(t *testing.T) {
	res := &HeadlessResult{
		ScriptName:  "demo.json",
		TotalNodes:  10,
		DialogCount: 5,
		ChoiceCount: 2,
		LabelCount:  1,
		SystemCount: 2,
	}
	s := res.String()
	if !strings.Contains(s, "demo.json") {
		t.Error("String() should contain script name")
	}
	if !strings.Contains(s, "dialogs=5") {
		t.Error("String() should contain dialog count")
	}
}
