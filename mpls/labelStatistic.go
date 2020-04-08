package mpls

// LabelStatistic refers to one entry in the stdout of `show mpls forwarding-table`.
type LabelStatistic struct {
	LocalLabel         string
	OutgoingLabel      string
	PrefixOrTunnelID   string
	OutgoingInterface  string
	NextHop            string
	BytesLabelSwitched float64
}

// MemoryStatistic refers to one entry in the stdout of `show mpls memory`.
type MemoryStatistic struct {
	AllocatorName string
	InUse         float64
	Allocated     float64
}
