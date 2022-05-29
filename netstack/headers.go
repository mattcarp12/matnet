package netstack

import "net"

type Header interface {
	Marshal() []byte
	Unmarshal([]byte) error
	GetType() ProtocolType
}

type L2Header interface {
	Header
	GetSrcMAC() net.HardwareAddr
	GetDstMAC() net.HardwareAddr
	GetL3Type() ProtocolType
}

type L3Header interface {
	Header
	GetSrcIP() net.IP
	GetDstIP() net.IP
	GetL4Type() ProtocolType
}

type L4Header interface {
	Header
	GetSrcPort() uint16
	GetDstPort() uint16
}

