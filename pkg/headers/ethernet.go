package headers

import (
	"errors"
	"net"
)

type EthernetHeader struct {
	DestinationMACAddress net.HardwareAddr
	SourceMACAddress      net.HardwareAddr
	EtherType             uint16
}

var ErrInvalidEthernetHeader = errors.New("invalid ethernet header")

func (eh *EthernetHeader) Unmarshal(b []byte) error {
	if len(b) < 14 {
		return ErrInvalidEthernetHeader
	}

	eh.DestinationMACAddress = net.HardwareAddr(b[0:6])
	eh.SourceMACAddress = net.HardwareAddr(b[6:12])
	eh.EtherType = uint16(b[12])<<8 | uint16(b[13]) // Is in big endian

	return nil
}
