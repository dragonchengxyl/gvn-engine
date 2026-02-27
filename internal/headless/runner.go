// Package headless provides script validation without any GUI dependency.
// It imports only compiler / history / script — zero ebiten dependency —
// so it can be built and tested in headless CI environments without X11.
package headless

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"gvn-engine/internal/compiler"
	"gvn-engine/internal/history"
	"gvn-engine/internal/script"
)

// Result summarises a headless script execution.
type Result struct {
	ScriptName  string
	TotalNodes  int
	DialogCount int
	ChoiceCount int
	LabelCount  int
	SystemCount int
	Warnings    []string
}

func (r *Result) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "[HEADLESS] script=%q  nodes=%d  dialogs=%d  choices=%d  labels=%d  system=%d",
		r.ScriptName, r.TotalNodes, r.DialogCount, r.ChoiceCount, r.LabelCount, r.SystemCount)
	for _, w := range r.Warnings {
		fmt.Fprintf(&sb, "\n  WARN: %s", w)
	}
	return sb.String()
}

// RunJSON walks a JSON script without rendering.
func RunJSON(name string, data []byte) (*Result, error) {
	sf, err := script.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("headless JSON: %w", err)
	}

	res := &Result{ScriptName: name, TotalNodes: len(sf.Commands)}
	state := history.GameState{
		Characters: make(map[string]history.CharSnapshot),
		Variables:  make(map[string]string),
	}

	for i, cmd := range sf.Commands {
		log.Printf("[HEADLESS] cmd[%03d] type=%-12s args=%v", i, cmd.Type, cmd.Args)
		switch cmd.Type {
		case script.CmdText:
			res.DialogCount++
			state.Speaker = cmd.Args["speaker"]
			state.FullText = cmd.Args["content"]
		case script.CmdBackground:
			state.Background = cmd.Args["file"]
			res.SystemCount++
		case script.CmdChar:
			if n := cmd.Args["name"]; n != "" {
				state.Characters[n] = history.CharSnapshot{
					File: cmd.Args["file"], Position: cmd.Args["position"],
				}
			}
			res.SystemCount++
		case script.CmdChoice:
			res.ChoiceCount++
		case script.CmdSet:
			if k := cmd.Args["key"]; k != "" {
				state.Variables[k] = cmd.Args["value"]
			}
			res.SystemCount++
		case script.CmdSound:
			if cmd.Args["mode"] == "bgm" {
				state.BGM = cmd.Args["file"]
			}
			res.SystemCount++
		case script.CmdFG:
			state.Foreground = cmd.Args["file"]
			res.SystemCount++
		case script.CmdJump, script.CmdIf, script.CmdWait:
			res.SystemCount++
		default:
			res.Warnings = append(res.Warnings,
				fmt.Sprintf("cmd[%d]: unhandled type %q", i, cmd.Type))
		}
	}

	// Validate labels
	for i, cmd := range sf.Commands {
		if cmd.Type == script.CmdJump {
			if t := cmd.Args["target"]; t != "" && sf.FindLabel(t) < 0 {
				res.Warnings = append(res.Warnings,
					fmt.Sprintf("cmd[%d]: jump target %q not found", i, t))
			}
		}
		if cmd.Type == script.CmdIf {
			if t := cmd.Args["jump"]; t != "" && sf.FindLabel(t) < 0 {
				res.Warnings = append(res.Warnings,
					fmt.Sprintf("cmd[%d]: if-true target %q not found", i, t))
			}
		}
	}
	_ = state
	return res, nil
}

// RunNVN compiles and walks a .nvn MD-VNDL script without rendering.
func RunNVN(name, src string) (*Result, error) {
	nodes, err := compiler.Compile(src)
	if err != nil {
		return nil, fmt.Errorf("headless NVN: %w", err)
	}

	res := &Result{ScriptName: name, TotalNodes: len(nodes)}
	state := history.GameState{
		Characters: make(map[string]history.CharSnapshot),
		Variables:  make(map[string]string),
	}

	for i, node := range nodes {
		switch n := node.(type) {
		case *compiler.LabelNode:
			res.LabelCount++
			log.Printf("[HEADLESS] node[%03d] LABEL  name=%s", i, n.Name)
		case *compiler.DialogNode:
			res.DialogCount++
			state.Speaker = n.Speaker
			state.FullText = n.Text
			log.Printf("[HEADLESS] node[%03d] DIALOG speaker=%q text=%q", i, n.Speaker, n.Text)
		case *compiler.ChoiceNode:
			res.ChoiceCount++
			log.Printf("[HEADLESS] node[%03d] CHOICE prompt=%q options=%d", i, n.Prompt, len(n.Options))
		case *compiler.SystemNode:
			res.SystemCount++
			applySystem(n, &state)
			log.Printf("[HEADLESS] node[%03d] SYSTEM cmd=%s val=%s opts=%v", i, n.Command, n.Value, n.Options)
		default:
			res.Warnings = append(res.Warnings,
				fmt.Sprintf("node[%d]: unknown type %T", i, node))
		}
	}
	_ = state
	return res, nil
}

// RunAuto detects format by extension (.nvn → NVN, else JSON).
func RunAuto(path string, data []byte) (*Result, error) {
	if strings.ToLower(filepath.Ext(path)) == ".nvn" {
		return RunNVN(path, string(data))
	}
	return RunJSON(path, data)
}

func applySystem(n *compiler.SystemNode, s *history.GameState) {
	switch n.Command {
	case "bg":
		s.Background = n.Value
	case "bgm":
		s.BGM = n.Value
	case "fg":
		s.Foreground = n.Value
	}
}
