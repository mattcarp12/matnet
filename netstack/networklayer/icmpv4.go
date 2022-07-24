package networklayer

import (
	"encoding/binary"
	"errors"
	"log"

	"github.com/mattcarp12/matnet/netstack"
)

type ICMPv4 struct {
	ip *IPv4
}

func NewICMPv4(ip *IPv4) *ICMPv4 {
	icmp := &ICMPv4{
		ip: ip,
	}

	return icmp
}

type ICMPv4Header struct {
	Type     uint8
	Code     uint8
	Checksum uint16
	Body     []byte
}

var ErrInvalidICMPHeader = errors.New("invalid ICMP header")

const (
	ICMPTypeEchoReply        = 0
	ICMPTypeEcho             = 8
	ICMPTypeDstUnreach       = 3
	ICMPTypeRedirect         = 5
	ICMPTypeTimeExceeded     = 11
	ICMPTypeParameterProblem = 12
)

// function to unmarshal the ICMP header
func (icmp *ICMPv4Header) Unmarshal(data []byte) error {
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
func (icmp *ICMPv4Header) Marshal() []byte {
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

	return b
}

func (icmp *ICMPv4) HandleRx(skb *netstack.SkBuff) {
	// Create a new ICMP header
	icmpHeader := &ICMPv4Header{}

	// Unmarshal the ICMP header
	if err := icmpHeader.Unmarshal(skb.Data); err != nil {
		return
	}

	// Handle the ICMP header
	switch icmpHeader.Type {
	case ICMPTypeEchoReply:
		log.Printf("ICMP echo reply") // TODO: Send to userspace?
	case ICMPTypeEcho:
		icmp.EchoReply(skb, icmpHeader)
	case ICMPTypeDstUnreach:
		icmp.HandleDestinationUnreachable(skb)
	case ICMPTypeRedirect:
		icmp.HandleRedirect(skb)
	case ICMPTypeTimeExceeded:
		log.Printf("ICMP Time Exceeded") // TODO: Send to userspace?
	default:
		log.Printf("ICMP Unknown")
	}
}

// function to handle ICMP echo requests
func (icmp *ICMPv4) EchoReply(skb *netstack.SkBuff, requestHeader *ICMPv4Header) {
	// Create new ICMP header
	icmpHeader := &ICMPv4Header{}

	// Fill in the ICMP header
	icmpHeader.Type = ICMPTypeEchoReply
	icmpHeader.Code = 0
	icmpHeader.Checksum = 0
	icmpHeader.Body = requestHeader.Body

	// Calculate the checksum
	rawHeader := icmpHeader.Marshal()
	icmpHeader.Checksum = netstack.Checksum(rawHeader)

	// Create new skb for ICMP echo reply
	rawReply := icmpHeader.Marshal()
	replySkb := netstack.NewSkBuff(rawReply)

	// Setup the reply skb
	txIface, err := skb.GetRxIface()
	if err != nil {
		return
	}

	replySkb.SetTxIface(txIface)
	replySkb.SetType(netstack.ProtocolTypeICMPv4)
	replySkb.SetSrcIP(skb.GetDstIP())
	replySkb.SetDstIP(skb.GetSrcIP())

	// Send the ICMP echo reply to IP
	icmp.ip.TxChan() <- replySkb

	// Make sure to read the skb response
	replySkb.GetResp()
}

// function to handle ICMP destination unreachable
func (icmp *ICMPv4) HandleDestinationUnreachable(skb *netstack.SkBuff) {
}

// function to handle ICMP redirect
func (icmp *ICMPv4) HandleRedirect(skb *netstack.SkBuff) {
	// Update routing table
}

// function to send PARAMETER PROBLEM message
func (icmp *ICMPv4) SendParamProblem(skb *netstack.SkBuff, code uint8) {
}

// function to send TIME EXCEEDED message
func (icmp *ICMPv4) SendTimeExceeded(skb *netstack.SkBuff, code uint8) {
}
