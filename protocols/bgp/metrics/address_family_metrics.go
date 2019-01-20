package metrics

type AddressFamilyMetrics struct {
	AFI             uint16
	SAFI            uint8
	UpdatesReceived uint64
	UpdatesSend     uint64
	RoutesReceived  uint64
	RoutesRejected  uint64
	RoutesAccepted  uint64
}
