package script

import (
	"encoding/json"
	"fmt"
	"log"
)

// validCommands is the set of recognized command types.
var validCommands = map[CommandType]bool{
	CmdBackground: true,
	CmdChar:       true,
	CmdText:       true,
	CmdChoice:     true,
	CmdSound:      true,
	CmdWait:       true,
	CmdShader:     true,
	CmdSet:        true,
	CmdIf:         true,
	CmdJump:       true,
	CmdFG:         true,
}

// Parse reads raw JSON bytes and returns a ScriptFile.
// Unknown or malformed commands are logged and skipped (Skip-on-Error Rule).
func Parse(data []byte) (*ScriptFile, error) {
	var raw struct {
		Title    string            `json:"title"`
		Commands []json.RawMessage `json:"commands"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("script: failed to parse JSON: %w", err)
	}

	sf := &ScriptFile{Title: raw.Title}

	for i, rawCmd := range raw.Commands {
		var cmd ScriptCommand
		if err := json.Unmarshal(rawCmd, &cmd); err != nil {
			log.Printf("[ERROR] script: line %d: malformed command, skipping: %v", i, err)
			continue
		}
		if !validCommands[cmd.Type] {
			log.Printf("[ERROR] script: line %d: unknown command type %q, skipping", i, cmd.Type)
			continue
		}
		sf.Commands = append(sf.Commands, cmd)
	}

	log.Printf("[INFO] script: loaded %q with %d commands", sf.Title, len(sf.Commands))
	sf.BuildLabels()
	return sf, nil
}
