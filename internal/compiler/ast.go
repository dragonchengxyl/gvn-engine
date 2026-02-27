package compiler

// NodeType identifies the kind of an AST node.
type NodeType int

const (
	NodeDialog  NodeType = iota // [角色] (params) "台词"
	NodeSystem                  // @command: value --key val
	NodeChoice                  // >> prompt:\n  - text -> Label
	NodeLabel                   // # LabelName
)

// Node is the interface all AST nodes implement.
type Node interface {
	Type() NodeType
}

// DialogNode represents a dialogue line.
// Example: [Sakura] (enter: left, expr: sad) "I don't want to do this."
type DialogNode struct {
	Speaker string            // character name
	Params  map[string]string // inline params: enter, expr, exit, etc.
	Text    string            // dialogue content
}

func (n *DialogNode) Type() NodeType { return NodeDialog }

// SystemNode represents a system directive.
// Example: @bg: school_day.png --fade 1.5
type SystemNode struct {
	Command string            // directive name: bg, bgm, se, fg, wait, set, jump, if
	Value   string            // primary value after colon
	Options map[string]string // --key value pairs
}

func (n *SystemNode) Type() NodeType { return NodeSystem }

// ChoiceOption is a single branch inside a ChoiceNode.
type ChoiceOption struct {
	Text  string // display text
	Label string // jump target label (may be empty)
}

// ChoiceNode represents a branching choice block.
// Example:
//
//	>> What will you do?
//	  - Fight -> Label_Fight
//	  - Run   -> Label_Flee
type ChoiceNode struct {
	Prompt  string         // optional prompt text after >>
	Options []ChoiceOption // ordered list of branches
}

func (n *ChoiceNode) Type() NodeType { return NodeChoice }

// LabelNode marks a jump target.
// Example: # Label_Fight
type LabelNode struct {
	Name string
}

func (n *LabelNode) Type() NodeType { return NodeLabel }
