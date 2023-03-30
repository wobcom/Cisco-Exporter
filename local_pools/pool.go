package local_pools

type Pool struct {
	Name string

	AddressesTotal    float64
	AddressesAvail    float64
	AddressesAssigned float64
}
