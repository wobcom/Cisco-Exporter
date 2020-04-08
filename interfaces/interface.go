package interfaces

// Interface represents a network interface on the remote device
type Interface struct {
	Name        string
	MacAddress  string
	Description string

	AdminStatus string
	OperStatus  string

	InputErrors  float64
	OutputErrors float64

	InputDrops  float64
	OutputDrops float64

	InputBytes  float64
	OutputBytes float64

	Speed string
}
