package relay

type Outputs [6]bool
type Inputs [7]bool
type Counters [7]uint16

type State struct {
	Outputs Outputs
	Inputs Inputs
	Clicks Counters
}

type StateChangedArgs struct {
	Old State
	New State
}