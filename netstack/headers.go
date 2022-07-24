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

// Checksum calculates the internet checksum
func Checksum(data []byte) uint16 {
	var (
		index  int
		sum    uint32
		length = len(data)
	)

	// make length multiple of 2
	if length%2 != 0 {
		data = append(data, 0)
		length++
	}

	// sum the 16-bit words
	for index < length {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
	}

	// add top 16 bits to bottom 16 bits
	sum = (sum >> 16) + (sum & 0xffff)

	// return 1's complement of sum
	return uint16(^sum)
}
