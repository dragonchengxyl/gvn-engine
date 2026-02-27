package history

import "log"

const maxStackSize = 200 // keep last N dialogue steps

// Stack is the Event Sourcing history stack.
// Apply pushes a new action; Undo walks back one step.
type Stack struct {
	state   GameState
	actions []Action
	cursor  int // points to the next free slot (0 = empty)
}

// NewStack creates a Stack initialised with a copy of the provided state.
func NewStack(initial GameState) *Stack {
	return &Stack{
		state:  initial,
		cursor: 0,
	}
}

// State returns a pointer to the current live GameState.
// Callers may read from it; mutations must go through Apply/Undo.
func (st *Stack) State() *GameState { return &st.state }

// Apply executes action.Apply on the current state and records it.
// If undo was used before Apply, any "future" redoable actions are discarded.
func (st *Stack) Apply(action Action) error {
	if err := action.Apply(&st.state); err != nil {
		return err
	}
	// Discard any future branch
	st.actions = st.actions[:st.cursor]
	st.actions = append(st.actions, action)
	st.cursor++

	// Trim oldest entries when we overflow
	if len(st.actions) > maxStackSize {
		trim := len(st.actions) - maxStackSize
		st.actions = st.actions[trim:]
		st.cursor = len(st.actions)
	}
	return nil
}

// Undo walks back one action and returns true.
// Returns false (and logs) when already at the oldest state.
func (st *Stack) Undo() bool {
	if st.cursor <= 0 {
		log.Printf("[INFO] history: already at oldest state")
		return false
	}
	st.cursor--
	if err := st.actions[st.cursor].Undo(&st.state); err != nil {
		log.Printf("[WARN] history: undo error: %v", err)
		st.cursor++ // roll back the cursor move
		return false
	}
	return true
}

// CanUndo reports whether there is at least one action to undo.
func (st *Stack) CanUndo() bool { return st.cursor > 0 }

// Len returns the total number of recorded actions.
func (st *Stack) Len() int { return st.cursor }
