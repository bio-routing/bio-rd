package packet

import (
	"bytes"
	"fmt"
)

const (
	// MinLen is the minimal length of a BMP message
	MinLen = 6

	RouteMonitoringType       = 0
	StatisticsReportType      = 1
	PeerDownNotificationType  = 2
	PeerUpNotificationType    = 3
	InitiationMessageType     = 4
	TerminationMessageType    = 5
	RouteMirroringMessageType = 6

	BGPMessage     = 0
	BGPInformation = 1

	ErroredPDU  = 0
	MessageLost = 1

	// BMPVersion is the supported BMP version
	BMPVersion = 3
)

// Msg is an interface that every BMP message must fulfill
type Msg interface {
	MsgType() uint8
}

// Decode decodes a BMP message
func Decode(msg []byte) (Msg, error) {
	buf := bytes.NewBuffer(msg)

	ch, err := decodeCommonHeader(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode common header: %v", err)
	}

	if ch.Version != BMPVersion {
		return nil, fmt.Errorf("Unsupported BMP version: %d", ch.Version)
	}

	switch ch.MsgType {
	case RouteMonitoringType:
		rm, err := decodeRouteMonitoringMsg(buf, ch)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode route monitoring message: %v", err)
		}

		return rm, err
	case StatisticsReportType:
		sr, err := decodeStatsReport(buf, ch)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode stats report: %v", err)
		}

		return sr, nil
	case PeerDownNotificationType:
		pd, err := decodePeerDownNotification(buf, ch)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode peer down notification: %v", err)
		}

		return pd, nil
	case PeerUpNotificationType:
		pu, err := decodePeerUpNotification(buf, ch)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode peer up notification: %v", err)
		}

		return pu, nil
	case InitiationMessageType:
		im, err := decodeInitiationMessage(buf, ch)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode initiation message: %v", err)
		}

		return im, nil
	case TerminationMessageType:
		tm, err := decodeTerminationMessage(buf, ch)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode termination message: %v", err)
		}

		return tm, nil
	case RouteMirroringMessageType:
		rm, err := decodeRouteMirroringMsg(buf, ch)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode route mirroring message: %v", err)
		}

		return rm, nil
	default:
		return nil, fmt.Errorf("Unexpected message type: %d", ch.MsgType)

	}
}
