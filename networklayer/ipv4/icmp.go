package ipv4

import (
	"encoding/binary"
	"errors"
	"log"

	"github.com/mattcarp12/go-net/netstack"
)

type ICMP struct {
}

type ICMPHeader struct {
	Type     uint8
	Code     uint8
	Checksum uint16
	Body     []byte
}

var ErrInvalidICMPHeader = errors.New("invalid ICMP header")

const (
	ICMPTypeEchoReply    = 0
	ICMPTypeEcho         = 8
	ICMPTypeDstUnreach   = 3
	ICMPTypeTimeExceeded = 11
)

// function to unmarshal the ICMP header
func (icmp *ICMPHeader) Unmarshal(data []byte) error {
	// check the length of the ICMP header
	if len(data) < 4 {
		return ErrInvalidICMPHeader
	}

	// type
	icmp.Type = data[0]

	// code
	icmp.Code = data[1]

	// checksum
	icmp.Checksum = binary.BigEndian.Uint16(data[2:4])

	// body
	icmp.Body = data[4:]

	return nil
}

// function to marshal the ICMP header
func (icmp *ICMPHeader) Marshal() ([]byte, error) {
	// make byte buffer for ICMP header
	b := make([]byte, 4)

	// type
	b[0] = icmp.Type

	// code
	b[1] = icmp.Code

	// checksum
	binary.BigEndian.PutUint16(b[2:4], icmp.Checksum)

	// body
	b = append(b, icmp.Body...)

	return b, nil
}

func (icmp *ICMP) HandleRx(skb *netstack.SkBuff) {
	// Create a new ICMP header
	icmpHeader := &ICMPHeader{}

	// Unmarshal the ICMP header
	err := icmpHeader.Unmarshal(skb.Data)
	if err != nil {
		log.Printf("Error unmarshalling ICMP header: %v", err)
		return
	}

	// Handle the ICMP packet

}

// function to handle ICMP echo requests
func (icmp *ICMP) EchoReply(skb *netstack.SkBuff) {

}
