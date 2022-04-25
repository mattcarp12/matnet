package protocols

import "errors"

type ProtocolType uint16

const (
	ProtocolTypeEthernet ProtocolType = iota
	ProtocolTypeIPv4
	ProtocolTypeICMPv4
	ProtocolTypeARP
	ProtocolTypeIPv6
	ProtocolTypeICMPv6
	ProtocolTypeTCP
	ProtocolTypeUDP
	ProtocolUnknown ProtocolType = 0xFFFF
)

var ErrProtocolNotFound = errors.New("protocol not found")
