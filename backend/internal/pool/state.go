package pool

// State is the lifecycle of one logical connection.
type State string

const (
	StateConnecting   State = "CONNECTING"
	StateIdle         State = "IDLE"
	StateInUse        State = "IN_USE"
	StateProbing      State = "PROBING"
	StateReconnecting State = "RECONNECTING"
	StateClosing      State = "CLOSING"
	StateClosed       State = "CLOSED"
)

func (s State) Borrowable() bool { return s == StateIdle }

func (s State) Live() bool {
	switch s {
	case StateClosed:
		return false
	default:
		return true
	}
}

func AllowedTransition(from, to State) bool {
	switch from {
	case StateConnecting:
		return to == StateIdle || to == StateInUse || to == StateClosing
	case StateIdle:
		return to == StateInUse || to == StateProbing || to == StateClosing
	case StateInUse:
		return to == StateIdle || to == StateReconnecting || to == StateClosing
	case StateProbing:
		return to == StateIdle || to == StateReconnecting || to == StateClosing
	case StateReconnecting:
		return to == StateIdle || to == StateClosing
	case StateClosing:
		return to == StateClosed
	default:
		return false
	}
}
