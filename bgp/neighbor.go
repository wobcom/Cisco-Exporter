package bgp

// Neighbor is a representation of the Cisco CLI ouputs concerning a BGP neigbor.
type Neighbor struct {
	RemoteAS    string
	RemoteIP    string
	Description string

	BGPVersion    float64
	State         string
	AdminShutdown float64

	HoldTime          float64
	KeepaliveInterval float64

	OpensSent         float64
	OpensRcvd         float64
	NotificationsSent float64
	NotificationsRcvd float64
	UpdatesSent       float64
	UpdatesRcvd       float64
	KeepalivesSent    float64
	KeepalivesRcvd    float64
	RouteRefreshsSent float64
	RouteRefreshsRcvd float64

	PrefixesCurrentBytes map[string]float64
	PrefixesCurrentRcvd  map[string]float64
	PrefixesCurrentSent  map[string]float64
	PrefixesTotalRcvd    map[string]float64
	PrefixesTotalSent    map[string]float64
	ImplicitWithdrawRcvd map[string]float64
	ImplicitWithdrawSent map[string]float64
	ExplicitWithdrawRcvd map[string]float64
	ExplicitWithdrawSent map[string]float64
	UsedAsBestpath       map[string]float64
	UsedAsMultipath      map[string]float64
	UsedAsSecondary      map[string]float64

	ConnectionsEstablished float64
	ConnectionsDropped     float64

	Uptime float64
}

// NewNeighbor returns a new bgp.Neighbor and initializes some fields
func NewNeighbor() *Neighbor {
	return &Neighbor{
		PrefixesCurrentBytes: make(map[string]float64),
		PrefixesCurrentRcvd:  make(map[string]float64),
		PrefixesCurrentSent:  make(map[string]float64),
		PrefixesTotalRcvd:    make(map[string]float64),
		PrefixesTotalSent:    make(map[string]float64),
		ImplicitWithdrawRcvd: make(map[string]float64),
		ImplicitWithdrawSent: make(map[string]float64),
		ExplicitWithdrawRcvd: make(map[string]float64),
		ExplicitWithdrawSent: make(map[string]float64),
		UsedAsBestpath:       make(map[string]float64),
		UsedAsMultipath:      make(map[string]float64),
		UsedAsSecondary:      make(map[string]float64),
	}
}
