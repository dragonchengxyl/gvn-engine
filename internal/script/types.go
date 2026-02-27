package script

// CommandType defines the type of a script command.
type CommandType string

const (
	CmdBackground CommandType = "bg"
	CmdChar       CommandType = "char"
	CmdText       CommandType = "text"
	CmdChoice     CommandType = "choice"
	CmdSound      CommandType = "sound"
	CmdWait       CommandType = "wait"
	CmdShader     CommandType = "shader"
	CmdSet        CommandType = "set"  // set a variable
	CmdIf         CommandType = "if"   // conditional jump
	CmdJump       CommandType = "jump" // unconditional jump
	CmdFG         CommandType = "fg"   // foreground overlay
)

// ScriptCommand represents a single instruction in the game script.
type ScriptCommand struct {
	Type     CommandType       `json:"type"`
	Args     map[string]string `json:"args,omitempty"`
	Duration float64           `json:"duration,omitempty"`
	Next     string            `json:"next,omitempty"`
}

// ScriptFile represents a complete script document.
type ScriptFile struct {
	Title    string          `json:"title"`
	Commands []ScriptCommand `json:"commands"`
	Labels   map[string]int  `json:"-"` // label -> command index (built after parse)
}

// BuildLabels scans commands and indexes their "next" fields as jump targets.
func (sf *ScriptFile) BuildLabels() {
	sf.Labels = make(map[string]int)
	for i, cmd := range sf.Commands {
		if cmd.Next != "" {
			sf.Labels[cmd.Next] = i
		}
	}
}

// FindLabel returns the command index for a label, or -1 if not found.
func (sf *ScriptFile) FindLabel(label string) int {
	if idx, ok := sf.Labels[label]; ok {
		return idx
	}
	return -1
}
