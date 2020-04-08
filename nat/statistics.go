package nat

// Statistics contains the information returned by a `show ip nat statistics`
type Statistics struct {
	ActiveTranslations        float64
	ActiveStaticTranslations  float64
	ActiveDynamicTranslations float64

	OutsideInterfaces   []string
	InsideInterfaces    []string
	Hits                float64
	Misses              float64
	ExpiredTranslations float64

	InToOutDrops float64
	OutToInDrops float64

	LimitMaxAllowed float64
	LimitUsed       float64
	LimitMissed     float64

	PoolStatsDrop      float64
	MappingStatsDrop   float64
	PortBlockAllocFail float64
	IPAliasAddFail     float64
	LimitEntryAddFail  float64
}

// NewStatistics returns a new Statistics Instance and initializes some fields
func NewStatistics() *Statistics {
	return &Statistics{
		OutsideInterfaces: make([]string, 0),
		InsideInterfaces:  make([]string, 0),
	}
}
