package aaa

// RadiusServer is a representation of the Cisco CLI outputs concerning a radius server.
type RadiusServer struct {
	ID             string
	Priority       string
	Host           string
	AuthPort       string
	AccountingPort string

	Up            float64
	UpDuration    float64
	DeadTotalTime float64
	DeadCount     float64

	Quarantined float64

	Requests        map[string]float64
	Timeouts        map[string]float64
	Failovers       map[string]float64
	Retransmissions map[string]float64

	Responses    map[string]map[string]float64
	ResponseTime map[string]float64

	SuccessfullTransactions map[string]float64
	FailedTransactions      map[string]float64
	ThrottledTransactions   map[string]float64
	ThrottledTimeouts       map[string]float64
	ThrottledFailures       map[string]float64
	MalformedResponses      map[string]float64
	BadAuthenticators       map[string]float64

	EstimatedOutstandingAccessTransactions     float64
	EstimatedOutstandingAccountingTransactions float64
	EstimatedThrottledAccessTransactions       float64
	EstimatedThrottledAccountingTransactions   float64
	RequestsPerMinuteHigh                      float64
	RequestsPerMinuteLow                       float64
	RequestsPerMinuteAverage                   float64
}

// NewRadiusServer returns a new instance of aaa.RadiusServer and initializes some fields.
func NewRadiusServer() *RadiusServer {
	return &RadiusServer{
		Requests:                make(map[string]float64),
		Timeouts:                make(map[string]float64),
		Failovers:               make(map[string]float64),
		Retransmissions:         make(map[string]float64),
		Responses:               make(map[string]map[string]float64),
		ResponseTime:            make(map[string]float64),
		SuccessfullTransactions: make(map[string]float64),
		FailedTransactions:      make(map[string]float64),
		ThrottledTransactions:   make(map[string]float64),
		ThrottledTimeouts:       make(map[string]float64),
		ThrottledFailures:       make(map[string]float64),
		MalformedResponses:      make(map[string]float64),
		BadAuthenticators:       make(map[string]float64),
	}
}
