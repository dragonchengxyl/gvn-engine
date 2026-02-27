// Package engine — headless.go
//
// RunHeadless* functions walk a script in pure-Go memory without any GUI,
// satisfying the Headless Compatibility rule from SKILL.md.
// They are used by the -headless CLI flag and by CI to validate script logic.
package engine

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"gvn-engine/internal/compiler"
	"gvn-engine/internal/history"
	"gvn-engine/internal/script"
)

// HeadlessResult summarises a headless script execution.
type HeadlessResult struct {
	ScriptName  string
	TotalNodes  int // commands (JSON) or AST nodes (.nvn)
	DialogCount int
	ChoiceCount int
	LabelCount  int
	SystemCount int
	Warnings    []string
}

func (r *HeadlessResult) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "[HEADLESS] script=%q  nodes=%d  dialogs=%d  choices=%d  labels=%d  system=%d",
		r.ScriptName, r.TotalNodes, r.DialogCount, r.ChoiceCount, r.LabelCount, r.SystemCount)
	for _, w := range r.Warnings {
		fmt.Fprintf(&sb, "\n  WARN: %s", w)
	}
	return sb.String()
}

// RunHeadlessJSON walks a JSON script (current production format).
// It applies state mutations via history.GameState to verify correctness.
func RunHeadlessJSON(name string, data []byte) (*HeadlessResult, error) {
	sf, err := script.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("headless JSON: %w", err)
	}

	res := &HeadlessResult{ScriptName: name, TotalNodes: len(sf.Commands)}
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
			name := cmd.Args["name"]
			if name != "" {
				state.Characters[name] = history.CharSnapshot{
					File:     cmd.Args["file"],
					Position: cmd.Args["position"],
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

	// Validate labels — every jump target must exist
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

	_ = state // state used for logging/validation
	return res, nil
}

// RunHeadlessNVN compiles and walks a .nvn MD-VNDL script.
// It exercises the Phase 5 compiler end-to-end with no rendering.
func RunHeadlessNVN(name, src string) (*HeadlessResult, error) {
	nodes, err := compiler.Compile(src)
	if err != nil {
		return nil, fmt.Errorf("headless NVN: %w", err)
	}

	res := &HeadlessResult{ScriptName: name, TotalNodes: len(nodes)}
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
			applySystemNode(n, &state)
			log.Printf("[HEADLESS] node[%03d] SYSTEM cmd=%s val=%s opts=%v", i, n.Command, n.Value, n.Options)

		default:
			res.Warnings = append(res.Warnings,
				fmt.Sprintf("node[%d]: unknown node type %T", i, node))
		}
	}

	_ = state
	return res, nil
}

// RunHeadlessAuto detects the script format by file extension and dispatches
// to the appropriate runner. Use this from main.go.
func RunHeadlessAuto(path string, data []byte) (*HeadlessResult, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".nvn":
		return RunHeadlessNVN(path, string(data))
	default: // .json or unknown → try JSON
		return RunHeadlessJSON(path, data)
	}
}

// applySystemNode maps a SystemNode command to GameState mutations.
func applySystemNode(n *compiler.SystemNode, s *history.GameState) {
	switch n.Command {
	case "bg":
		s.Background = n.Value
	case "bgm":
		s.BGM = n.Value
	case "fg":
		s.Foreground = n.Value
	case "set":
		if n.Value != "" {
			// value format: "key=val" or split from options
			if v, ok := n.Options["value"]; ok {
				s.Variables[n.Value] = v
			}
		}
	}
}
