package protocols

type NetworkProtocolID uint8

const (
	IPv4 NetworkProtocolID = iota
	IPv6
	ARP
)

func L3_Protocol_From_EtherType(etherType uint16) NetworkProtocolID {
	switch etherType {
	case 0x0800:
		return IPv4
	case 0x86DD:
		return IPv6
	case 0x0806:
		return ARP
	default:
		return 0
	}
}
