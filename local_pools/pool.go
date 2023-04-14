package local_pools

type PoolGroup struct {
	Name string

	Pools []Pool
}

type Pool struct {
	StartIP string
	EndIP   string

	AddressesTotal    float64
	AddressesAvail    float64
	AddressesAssigned float64
}
