package nat

// Pool represents one NAT pool returned by `show ip nat pool name "..."`
type Pool struct {
	Name string
	ID   string

	Type string

	Refcount float64

	Misses float64

	Netmask string
	StartIP string
	EndIP   string

	AddressesTotal    float64
	AddressesAvail    float64
	AddressesAssigned float64

	// Low ports are less than 1024. High ports are greater than or equal to 1024.
	UDPLowPortsAvail    float64
	UDPLowPortsAssigned float64

	TCPLowPortsAvail    float64
	TCPLowPortsAssigned float64

	UDPHighPortsAvail    float64
	UDPHighPortsAssigned float64

	TCPHighPortsAvail    float64
	TCPHighPortsAssigned float64
}
