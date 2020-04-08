package opticsxe

// XETransceiver represent a transceiver inserted into a cisco device running IOS XE.
type XETransceiver struct {
	Slot          string
	Subslot       string
	Port          string
	Enabled       bool
	Temperature   float64
	BiasCurrent   float64
	TransmitPower float64
	ReceivePower  float64
}
