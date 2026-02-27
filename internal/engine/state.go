package engine

// EngineState represents the current state of the game engine.
type EngineState int

const (
	StateLoading    EngineState = iota // Loading resources
	StateIdle                          // Ready to process next command
	StateTyping                        // Text is being typed out
	StateWaitInput                     // Waiting for player click/tap
	StateTransition                    // Playing a transition animation
	StateChoice                        // Displaying choice options
	StateMenu                          // In settings/save menu
)

func (s EngineState) String() string {
	switch s {
	case StateLoading:
		return "Loading"
	case StateIdle:
		return "Idle"
	case StateTyping:
		return "Typing"
	case StateWaitInput:
		return "WaitInput"
	case StateTransition:
		return "Transition"
	case StateChoice:
		return "Choice"
	case StateMenu:
		return "Menu"
	default:
		return "Unknown"
	}
}
