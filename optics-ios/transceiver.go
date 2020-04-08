package opticsios

// Transceiver represent a transceiver inserted into a cisco device running IOS.
type Transceiver struct {
	Name          string
	Temperature   map[string]float64
	Voltage       map[string]float64
	Current       map[string]float64
	TransmitPower map[string]float64
	ReceivePower  map[string]float64
}

// NewTransceiver initializes some fields and returns a new optics.Transceiver instance.
func NewTransceiver() *Transceiver {
	return &Transceiver{
		Temperature:   make(map[string]float64),
		Voltage:       make(map[string]float64),
		Current:       make(map[string]float64),
		TransmitPower: make(map[string]float64),
		ReceivePower:  make(map[string]float64),
	}
}
