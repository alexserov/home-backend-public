package relay

type outputs [6]bool
type inputs [7]bool
type counters [7]uint16

type state struct {
	outputs outputs
	inputs inputs
	counters counters
}