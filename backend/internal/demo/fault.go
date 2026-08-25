package demo

type FaultState struct {
	FailPing bool `json:"fail_ping"`
	FailDial bool `json:"fail_dial"`
	DropNext int  `json:"drop_next"`
}

type FaultSink interface {
	SetFailPing(bool)
	SetFailDial(bool)
	SetDropNext(int)
	FaultState() (bool, bool, int)
}

func ApplyFaults(sink FaultSink, failPing, failDial *bool, drop *int) FaultState {
	if sink == nil {
		return FaultState{}
	}
	if failPing != nil {
		sink.SetFailPing(*failPing)
	}
	if failDial != nil {
		sink.SetFailDial(*failDial)
	}
	if drop != nil {
		sink.SetDropNext(*drop)
	}
	a, b, n := sink.FaultState()
	return FaultState{FailPing: a, FailDial: b, DropNext: n}
}
