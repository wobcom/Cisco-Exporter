package opticsnxos

// NXOSTransceiver represents a transceiver inserted into a cisco device running NX OS.
type NXOSTransceiver struct {
	Lane          string
	Name          string
	Temperature   map[string]float64
	Voltage       map[string]float64
	Current       map[string]float64
	TransmitPower map[string]float64
	ReceivePower  map[string]float64
	Faultcount    float64
}

// NewTransceiver initializes some fields and returns a new opticsnxos.NXOSTransceiver intance.
func NewTransceiver() *NXOSTransceiver {
	return &NXOSTransceiver{
		Temperature:   make(map[string]float64),
		Voltage:       make(map[string]float64),
		Current:       make(map[string]float64),
		TransmitPower: make(map[string]float64),
		ReceivePower:  make(map[string]float64),
	}
}
