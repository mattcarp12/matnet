package transportlayer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/mattcarp12/matnet/netstack"
)

// =============================================================================
// TCP Header
// =============================================================================
type TCPHeader struct {
	SrcPort   uint16
	DstPort   uint16
	SeqNum    uint32
	AckNum    uint32
	HeaderLen uint8 // 4 bits + reserved
	BitFlags  uint8 // 8 bits
	Window    uint16
	Checksum  uint16
	UrgentPtr uint16
	Options   TCPOptions
}

const (
	TCP_FIN = 0x01
	TCP_SYN = 0x02
	TCP_RST = 0x04
	TCP_PSH = 0x08
	TCP_ACK = 0x10
	TCP_URG = 0x20
	TCP_ECE = 0x40
	TCP_CWR = 0x80
)

const (
	TCPHeaderMinSize = 20
	TCPWordSize      = 4
)

func (h TCPHeader) Marshal() []byte {
	b := make([]byte, TCPHeaderMinSize)
	binary.BigEndian.PutUint16(b[0:2], h.SrcPort)
	binary.BigEndian.PutUint16(b[2:4], h.DstPort)
	binary.BigEndian.PutUint32(b[4:8], h.SeqNum)
	binary.BigEndian.PutUint32(b[8:12], h.AckNum)
	headerLen := h.HeaderLen << 4
	copy(b[12:14], []byte{headerLen, h.BitFlags})
	binary.BigEndian.PutUint16(b[14:16], h.Window)
	binary.BigEndian.PutUint16(b[16:18], h.Checksum)
	binary.BigEndian.PutUint16(b[18:20], h.UrgentPtr)

	if h.Options != nil {
		b = append(b, h.Options.Marshal()...)
	}

	return b
}

func (h *TCPHeader) Unmarshal(b []byte) error {
	h.SrcPort = binary.BigEndian.Uint16(b[0:2])
	h.DstPort = binary.BigEndian.Uint16(b[2:4])
	h.SeqNum = binary.BigEndian.Uint32(b[4:8])
	h.AckNum = binary.BigEndian.Uint32(b[8:12])
	h.HeaderLen = b[12] >> 4
	h.BitFlags = b[13]
	h.Window = binary.BigEndian.Uint16(b[14:16])
	h.Checksum = binary.BigEndian.Uint16(b[16:18])
	h.UrgentPtr = binary.BigEndian.Uint16(b[18:20])

	if h.SizeInBytes() > TCPHeaderMinSize {
		h.Options = TCPOptions{}
		optionsBytes := b[20:h.SizeInBytes()]
		h.Options.Unmarshal(optionsBytes)
	}

	return nil
}

func (h TCPHeader) SizeInBytes() int {
	return int(h.HeaderLen * TCPWordSize)
}

// Implement the netstack.L4Header interface
func (h TCPHeader) GetSrcPort() uint16 {
	return h.SrcPort
}

func (h TCPHeader) GetDstPort() uint16 {
	return h.DstPort
}

func (h TCPHeader) GetType() netstack.ProtocolType {
	return netstack.ProtocolTypeTCP
}

func (h TCPHeader) IsFIN() bool {
	return h.BitFlags&TCP_FIN == TCP_FIN
}

func (h *TCPHeader) SetFIN() {
	h.BitFlags |= TCP_FIN
}

func (h TCPHeader) IsSYN() bool {
	return h.BitFlags&TCP_SYN == TCP_SYN
}

func (h *TCPHeader) SetSYN() {
	h.BitFlags |= TCP_SYN
}

func (h TCPHeader) IsRST() bool {
	return h.BitFlags&TCP_RST == TCP_RST
}

func (h *TCPHeader) SetRST() {
	h.BitFlags |= TCP_RST
}

func (h TCPHeader) IsPSH() bool {
	return h.BitFlags&TCP_PSH == TCP_PSH
}

func (h *TCPHeader) SetPSH() {
	h.BitFlags |= TCP_PSH
}

func (h TCPHeader) IsACK() bool {
	return h.BitFlags&TCP_ACK == TCP_ACK
}

func (h *TCPHeader) SetACK() {
	h.BitFlags |= TCP_ACK
}

type TCPOptionKind uint8

const (
	TCPOptionKindEndOfOptions TCPOptionKind = 0
	TCPOptionKindNop                        = 1
	TCPOptionKindMSS                        = 2
	TCPOptionKindWS                         = 3
)

type TCPOption struct {
	Kind      TCPOptionKind
	Mss       uint16
	Nop       uint8
	Nop2      uint8
	WScale    uint8
	SackPerm  uint8
	Sack      uint8
	Timestamp uint8
}

type TCPOptions []TCPOption

func (o TCPOptions) Marshal() []byte {
	b := make([]byte, 0)
	for _, option := range o {
		b = append(b, option.Marshal()...)
	}
	return b
}

func (o TCPOption) Marshal() []byte {
	b := make([]byte, 12)
	binary.BigEndian.PutUint16(b[0:2], o.Mss)
	b[2] = o.Nop
	b[3] = o.Nop2
	b[4] = o.WScale
	b[5] = o.SackPerm
	b[6] = o.Sack
	b[7] = o.Timestamp

	return b
}

func (o *TCPOptions) Unmarshal(b []byte) {
	remainingLength := len(b)
	optionLength := 0

	for {
		optionKind := TCPOptionKind(b[0])
		option := TCPOption{}
		option.Kind = optionKind

		switch optionKind {
		case TCPOptionKindEndOfOptions, TCPOptionKindNop:
			// End of options
			return
		case 2:
			// MSS
			option.Mss = binary.BigEndian.Uint16(b[2:4])
			optionLength = int(b[1])

		default:
			fmt.Printf("Unknown option kind: %d\n", optionKind)
		}

		b = b[optionLength:]
		remainingLength -= optionLength

		*o = append(*o, option)

		if remainingLength == 0 {
			return
		}
	}
}

type TCPPseudoHeader struct {
	SrcIP    net.IP
	DstIP    net.IP
	Zero     uint8
	Protocol uint8
	Length   uint16
}

func (ph TCPPseudoHeader) Marshal() []byte {
	b := []byte{}
	b = append(b, ph.SrcIP...)
	b = append(b, ph.DstIP...)
	b = append(b, ph.Zero)
	b = append(b, ph.Protocol)
	b = append(b, byte(ph.Length>>8))
	b = append(b, byte(ph.Length))

	return b
}

// ============================================================================
// TCP Control Block
// ============================================================================

type TCPState int

const (

	// Represents no connection state at all.
	TCP_STATE_CLOSED TCPState = iota

	// Represents waiting for a connection request.
	TCP_STATE_LISTEN

	// Represents waiting for a matching connection request after
	// sending a connection request.
	TCP_STATE_SYN_SENT

	// Represents waiting for a confirming connection request acknowledgment
	// after having both received and sent a connection request.
	TCP_STATE_SYN_RCVD

	// Represents a fully established connection. This is the normal state for data transfer.
	TCP_STATE_ESTABLISHED

	// Represents waiting for a connection termination request from the remote TCP,
	// or an acknowledgment of the connection termination request previously sent.
	TCP_STATE_FIN_WAIT_1

	// Represents waiting for a connection termination request from the remote TCP.
	TCP_STATE_FIN_WAIT_2

	// Represents waiting for a connection termination request from the local user.
	TCP_STATE_CLOSE_WAIT

	// Represnts waiting for a connection termination request acknowledgment from the remote TCP.
	TCP_STATE_CLOSING

	// Represents waiting for an acknowledgement of the connection termination request previously
	// sent to the remote TCP.
	TCP_STATE_LAST_ACK

	// Represents waiting for enough time to pass to be sure the remote TCP has received the
	// connection termination request acknowledgment.
	TCP_STATE_TIME_WAIT
)

type TCB struct {
	ID    string
	State TCPState

	SendUNA uint32 // Unacknowledged sequence number
	SendNXT uint32 // Next sequence number to send
	SendWND uint32 // Window size
	SendUP  uint32 // Urgent pointer
	SendWL1 uint32 // Seq num used for last window update
	SendWL2 uint32 // Ack num used for last window update
	SendISN uint32 // Initial sequence number

	RecvNXT uint32 // Next sequence number to receive
	RecvWND uint32 // Window size
	RecvUP  uint32 // Urgent pointer
	RecvISN uint32 // Initial sequence number

	RxChan       chan *netstack.SkBuff
	RxChanSorted chan *netstack.SkBuff
	RxQueue      SegmentQueue
}

func (tcp *TCPProtocol) NewTCB(connID string) *TCB {
	tcb := &TCB{
		ID:           connID,
		State:        TCP_STATE_CLOSED,
		RxChan:       make(chan *netstack.SkBuff),
		RxChanSorted: make(chan *netstack.SkBuff),
	}

	tcp.ConnTable[connID] = tcb

	return tcb
}

type SegmentQueue []*netstack.SkBuff

// func (q SegmentQueue) Len() int { return len(q) }
// func (q SegmentQueue) Less(i, j int) bool {
// 	segment1 := q[i].L4Header.(*TcpHeader)
// 	segment2 := q[j].L4Header.(*TcpHeader)
// 	return segment1.SeqNum < segment2.SeqNum
// }

// ============================================================================
// TCP Protocol
// ============================================================================
type TCPProtocol struct {
	netstack.IProtocol
	ConnTable map[string]*TCB
}

var (
	ErrInvalidSequenceNumber = errors.New("invalid sequence number")
	ErrInvalidAckNumber      = errors.New("invalid ack number")
	ErrAckNotSet             = errors.New("ack bit not set in header")
	ErrConnectionReset       = errors.New("connection reset")
)

func NewTCP() *TCPProtocol {
	tcp := &TCPProtocol{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeTCP),
		ConnTable: make(map[string]*TCB),
	}
	tcp.Log = netstack.NewLogger("TCP")

	return tcp
}

/*
	TCP HandleRx algorithm
	- Unmarshal the TCP header
	- Find the TCB. Create new TCB is packet is a SYN.
	- Check the sequence numbers, make sure packet is within the window.
	- Put the packet into the segment processing queue.
*/

func (tcp *TCPProtocol) HandleRx(skb *netstack.SkBuff) {
	tcp.Log.Printf("HandleRx: %+v\n", skb)
	// Create a new TCP header
	tcpHeader := &TCPHeader{}

	// Unmarshal the TCP header, handle errors
	if err := tcpHeader.Unmarshal(skb.Data); err != nil {
		skb.Error(err)
		return
	}

	skb.StripBytes(tcpHeader.SizeInBytes())
	skb.SetSrcPort(tcpHeader.SrcPort)
	skb.SetDstPort(tcpHeader.DstPort)

	// TODO: Handle fragmentation, possibly reassemble

	// Find the TCB for this connection. Here, LocalAddr = DstAddr, RemoteAddr = SrcAddr.
	connID := ConnectionID(skb.GetDstAddr(), skb.GetSrcAddr())

	var tcb *TCB

	tcb, ok := tcp.ConnTable[connID]
	if !ok {
		// TCB does not exist. All data is discarded.
		if tcpHeader.IsRST() {
			skb.Error(errors.New("TCP: RST received for non-existent connection"))
			return
		}

		tcp.SendEmptyRst(tcpHeader)
	}

	tcp.Log.Printf("TCP Header: %+v\n", tcpHeader)
	tcp.Log.Printf("TCB: %+v\n", tcb)
	tcp.Log.Printf("Header IsSyn: %v\n", tcpHeader.IsSYN())
	tcp.Log.Printf("Header IsACK: %v\n", tcpHeader.IsACK())

	// Handle the TCP packet

	// Check the sequence number, make sure it's in the window
	// if tcpHeader.SeqNum < tcb.RecvNXT || tcpHeader.SeqNum > tcb.RecvNXT+tcb.RecvWND {
	// 	// skb.Error(errors.New("TCP: SeqNum out of window"))
	// 	tcp.Log.Printf("TCP: SeqNum out of window\n")
	// 	return
	// }

	switch tcb.State {
	default:
		skb.Error(errors.New("TCP: Invalid state"))
	case TCP_STATE_LISTEN:
		if tcpHeader.IsSYN() {
			tcp.SendSynAck(skb, tcb, tcpHeader)
			return
		}
	case TCP_STATE_SYN_RCVD:
		// Here we sent a SYN+ACK, and now we are waiting for an ACK.
		if tcpHeader.IsACK() && tcpHeader.AckNum == tcb.SendNXT {
			tcb.State = TCP_STATE_ESTABLISHED
		}
	case TCP_STATE_SYN_SENT:
		if err := tcp.HandleSynSent(skb, tcb, tcpHeader); err != nil {
			skb.Error(err)
			return
		}

	}

	// Update the TCB

	// Everything is good, send to user space
}

func (tcp *TCPProtocol) HandleTx(skb *netstack.SkBuff) {
	// Create new TCP header

	// Calculate checksum

	// Set the skb's L4 header to the TCP header

	// Prepend the TCP header to the skb

	// Passing to the network layer, so set the skb type
	// to the type of the socket address family (ipv4, ipv6, etc)

	// Send to the network layer
	tcp.TxDown(skb)
}

func ISN() uint32 {
	// TODO: Find a better way to do this

	// Generate a random ISN
	rand.Seed(time.Now().UnixNano())
	return rand.Uint32()
}

// ==============================================================================
// TCP Event Handlers
// ==============================================================================

// SendSynAck sends a SYN/ACK packet to the remote TCP in response
// to a SYN packet. The TCB is updated with the remote TCP's ISN.
// The TCBs state is set to TCP_STATE_SYN_RCVD.
func (tcp *TCPProtocol) SendSynAck(skb *netstack.SkBuff, tcb *TCB, requestHeader *TCPHeader) {
	// Create a new TCP header
	newHeader := &TCPHeader{}

	// Set the TCP header fields
	newHeader.SrcPort = requestHeader.GetDstPort()
	newHeader.DstPort = requestHeader.GetSrcPort()
	newHeader.BitFlags = TCP_SYN | TCP_ACK
	newHeader.SeqNum = ISN()
	newHeader.AckNum = requestHeader.SeqNum + 1
	newHeader.HeaderLen = 5
	newHeader.Window = 0xFFFF
	newHeader.UrgentPtr = 0

	// Update the TCB
	tcb.SendISN = newHeader.SeqNum
	tcb.SendNXT = newHeader.SeqNum + 1
	tcb.SendWND = 0xFFFF
	tcb.SendUP = 0
	tcb.SendWL1 = newHeader.SeqNum
	tcb.SendWL2 = newHeader.SeqNum
	tcb.State = TCP_STATE_SYN_RCVD
	tcb.RecvISN = requestHeader.SeqNum
	tcb.RecvNXT = requestHeader.SeqNum + 1
	tcb.RecvWND = 0xFFFF
	tcb.RecvUP = 0

	// Set the skb's L4 header to the TCP header
	newSkb := netstack.NewSkBuff([]byte{})

	newSkb.SetSrcIP(skb.GetDstIP())
	newSkb.SetDstIP(skb.GetSrcIP())
	newSkb.SetSrcPort(skb.GetDstPort())
	newSkb.SetDstPort(skb.GetSrcPort())

	setSkbType(newSkb)

	// Send to the network layer
	tcp.TxDown(newSkb)

	// Wait for skb to be sent
	newSkb.GetResp()
}

func (tcp *TCPProtocol) SendAck(skb *netstack.SkBuff, tcb *TCB, requestHeader *TCPHeader) {
	tcp.Log.Printf("Sending Ack for %+v\n", requestHeader)
	newHeader := &TCPHeader{}

	newHeader.SrcPort = requestHeader.GetDstPort()
	newHeader.DstPort = requestHeader.GetSrcPort()
	newHeader.BitFlags = TCP_ACK
	newHeader.SeqNum = tcb.SendNXT
	newHeader.AckNum = tcb.RecvNXT
	newHeader.HeaderLen = 5
	newHeader.Window = 0xFFFF
	newHeader.UrgentPtr = 0

	newSkb := netstack.NewSkBuff([]byte{})
	newSkb.SetDstIP(skb.GetSrcIP())
	newSkb.SetSrcIP(skb.GetDstIP())
	newSkb.SetSrcPort(newHeader.SrcPort)
	newSkb.SetDstPort(newHeader.DstPort)

	setTCPChecksum(newSkb, newHeader)

	newSkb.SetL4Header(newHeader)
	newSkb.PrependBytes(newHeader.Marshal())

	rxIface, err := skb.GetRxIface()
	if err != nil {
		tcp.Log.Printf("Error getting rxIface: %v\n", err)
		return
	}

	newSkb.SetTxIface(rxIface)

	setSkbType(newSkb)

	tcp.TxDown(newSkb)

	newSkb.GetResp()
}

func (tcp *TCPProtocol) SendEmptyRst(tcpHeader *TCPHeader) {
	// Create a new TCP header
	newHeader := &TCPHeader{}

	// Set the TCP header fields
	newHeader.SrcPort = tcpHeader.GetDstPort()
	newHeader.DstPort = tcpHeader.GetSrcPort()
	newHeader.BitFlags = TCP_RST
	newHeader.AckNum = tcpHeader.SeqNum + 1
	newHeader.HeaderLen = 5
	newHeader.Window = 0xFFFF
	newHeader.UrgentPtr = 0
	if tcpHeader.IsACK() {
		newHeader.SeqNum = tcpHeader.AckNum
	} else {
		newHeader.SeqNum = 0
	}

	// Create a new skb
	newSkb := netstack.NewSkBuff([]byte{})

	// Send to the network layer
	tcp.TxDown(newSkb)
}

// OpenConnection...
// This function is meant to be called by the Socket layer,
// in response to a socket open request.
//
// This function sends a SYN packet to the remote TCP
// and returns immediately.
func (tcp *TCPProtocol) OpenConnection(srcAddr, dstAddr netstack.SockAddr, iface netstack.NetworkInterface) error {
	tcp.Log.Printf("OpenConnection: %v -> %v\n", srcAddr, dstAddr)

	// Make empty skb
	skb := netstack.NewSkBuff([]byte{})

	skb.SetSrcAddr(srcAddr)
	skb.SetDstAddr(dstAddr)
	skb.SetTxIface(iface)

	if err := setSkbType(skb); err != nil {
		return err
	}

	// Make TCP header
	header := &TCPHeader{}
	isn := ISN()

	header.SrcPort = srcAddr.Port
	header.DstPort = dstAddr.Port
	header.BitFlags = TCP_SYN
	header.SeqNum = isn
	header.AckNum = 0
	header.HeaderLen = 5
	header.Window = 0xFFFF
	header.UrgentPtr = 0

	setTCPChecksum(skb, header)
	skb.SetL4Header(header)
	skb.PrependBytes(header.Marshal())

	// Send to the network layer
	tcp.TxDown(skb)

	if skbResp := skb.GetResp(); skbResp.Error != nil {
		tcp.Log.Printf("OpenConnection: Error sending SYN: %v\n", skbResp.Error)
		return skbResp.Error
	}

	// Create a new TCB
	connID := ConnectionID(srcAddr, dstAddr)
	tcb := tcp.NewTCB(connID)
	tcb.State = TCP_STATE_SYN_SENT
	tcb.SendUNA = isn
	tcb.SendNXT = isn + 1

	return nil
}

func setTCPChecksum(skb *netstack.SkBuff, header *TCPHeader) {
	// Set the checksum initially to 0
	header.Checksum = 0

	// Make pseudo header
	// TODO: Handle IPv6
	ph := &TCPPseudoHeader{
		SrcIP:    skb.GetSrcIP(),
		DstIP:    skb.GetDstIP(),
		Zero:     0,
		Protocol: 6,
		Length:   uint16(header.HeaderLen*4) + uint16(len(skb.Data)),
	}

	// Create checksum buffer
	b := append(header.Marshal(), skb.Data...)
	b = append(ph.Marshal(), b...)

	// Calculate checksum
	header.Checksum = netstack.Checksum(b)
}

func (tcp *TCPProtocol) CloseConnection() {
	// This function is meant to be called by the Socket layer,
	// in response to a socket close request.

	// Find the TCB for this connection
	// Update the TCB
	// Send a FIN packet to the remote TCP
	// Wait for an ACK
}

// This function is called when a packet is received and the state is TCP_STATE_SYN_SENT
func (tcp *TCPProtocol) HandleSynSent(skb *netstack.SkBuff, tcb *TCB, tcpHeader *TCPHeader) error {
	tcp.Log.Printf("HandleSynSent: %v\n", tcpHeader)
	// First check the ACK bit
	if tcpHeader.IsACK() {
		if tcpHeader.AckNum <= tcb.SendISN || tcpHeader.AckNum > tcb.SendNXT {
			tcp.SendEmptyRst(tcpHeader)
			return fmt.Errorf("HandleSynSent: %w", ErrInvalidSequenceNumber)
		}
	}

	// Check if the ACK number is valid
	if tcpHeader.AckNum < tcb.SendUNA && tcpHeader.AckNum > tcb.SendNXT {
		return fmt.Errorf("HandleSynSent: %w", ErrInvalidAckNumber)
	}

	// Second check the RST bit
	if tcpHeader.IsRST() {
		// TODO: How to signal to the user that the connection was reset?

		// Remove the TCB
		delete(tcp.ConnTable, tcb.ID)

		return ErrConnectionReset
	}

	// TODO: Check Security and Precedence bits

	// Fourth check the SYN bit
	if tcpHeader.IsSYN() {
		tcb.RecvNXT = tcpHeader.SeqNum + 1
		tcb.RecvISN = tcpHeader.SeqNum
		tcb.SendUNA = tcpHeader.AckNum

		// Update the state
		if tcb.SendUNA > tcb.SendISN {
			tcb.State = TCP_STATE_ESTABLISHED
			tcp.SendAck(skb, tcb, tcpHeader)
		} else {
			tcb.State = TCP_STATE_SYN_RCVD
			tcp.SendSynAck(skb, tcb, tcpHeader)
		}
	}

	return nil
}

// ==============================================================================
// Sequence Number Functions
// ==============================================================================

func IsLessThan(isn, seq1, seq2 uint32) bool {
	// Compare using modular arithmetic
	// return (seq1-isn)%0xffffffff < (seq2-isn)%0xffffffff
	return seq1 < seq2
}
