package relay

type outputs struct {
	O1 bool
	O2 bool
	O3 bool
	O4 bool
	O5 bool
	O6 bool
}

type inputs struct {
	I0 bool
	I1 bool
	I2 bool
	I3 bool
	I4 bool
	I5 bool
	I6 bool
}

type counters struct {
	I0 uint16
	I1 uint16
	I2 uint16
	I3 uint16
	I4 uint16
	I5 uint16
	I6 uint16
}

type state struct {
	outputs outputs
	inputs inputs
	counters counters
}