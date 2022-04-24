package headers

import "net"

type IPv4Header struct {
	Version        uint8
	IHL            uint8
	TypeOfService  uint8
	TotalLength    uint16
	Identification uint16
	Flags          uint8
	FragmentOffset uint16
	TTL            uint8
	Protocol       uint8
	HeaderChecksum uint16
	SourceIP       net.IP
	DestinationIP  net.IP
}

type ARPHeader struct {
	HardwareType uint16
	ProtocolType uint16
	HardwareSize uint8
	ProtocolSize uint8
	OpCode       uint16
	SenderMAC    net.HardwareAddr
	SenderIP     net.IP
	TargetMAC    net.HardwareAddr
	TargetIP     net.IP
}
