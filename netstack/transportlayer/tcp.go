package transportlayer

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"net"
	"time"

	"github.com/mattcarp12/go-net/netstack"
)

/*******************************************************************************
	TCP Header
*******************************************************************************/
type TcpHeader struct {
	SrcPort   uint16
	DstPort   uint16
	SeqNum    uint32
	AckNum    uint32
	HeaderLen uint8 // 4 bits + reserved
	BitFlags  uint8 // 8 bits
	Window    uint16
	Checksum  uint16
	UrgentPtr uint16
	Options   *TcpOptions
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

func (h TcpHeader) Marshal() []byte {
	b := make([]byte, 20)
	binary.BigEndian.PutUint16(b[0:2], h.SrcPort)
	binary.BigEndian.PutUint16(b[2:4], h.DstPort)
	binary.BigEndian.PutUint32(b[4:8], h.SeqNum)
	binary.BigEndian.PutUint32(b[8:12], h.AckNum)
	copy(b[12:14], []byte{byte(h.HeaderLen), byte(h.BitFlags)})
	binary.BigEndian.PutUint16(b[14:16], h.Window)
	binary.BigEndian.PutUint16(b[16:18], h.Checksum)
	binary.BigEndian.PutUint16(b[18:20], h.UrgentPtr)

	if h.Options != nil {
		b = append(b, h.Options.Marshal()...)
	}

	return b
}

func (h *TcpHeader) Unmarshal(b []byte) error {
	h.SrcPort = binary.BigEndian.Uint16(b[0:2])
	h.DstPort = binary.BigEndian.Uint16(b[2:4])
	h.SeqNum = binary.BigEndian.Uint32(b[4:8])
	h.AckNum = binary.BigEndian.Uint32(b[8:12])
	h.HeaderLen = uint8(b[12]) >> 4
	h.BitFlags = uint8(b[12]) & 0x0f
	h.Window = binary.BigEndian.Uint16(b[16:18])
	h.Checksum = binary.BigEndian.Uint16(b[18:20])
	h.UrgentPtr = binary.BigEndian.Uint16(b[20:22])

	if h.HeaderLen > 5 {
		h.Options = &TcpOptions{}
		h.Options.Unmarshal(b[22 : h.HeaderLen*4])
	}

	return nil
}

func (h TcpHeader) Size() int {
	return int(h.HeaderLen * 4)
}

// Implement the netstack.L4Header interface
func (h TcpHeader) GetSrcPort() uint16 {
	return h.SrcPort
}

func (h TcpHeader) GetDstPort() uint16 {
	return h.DstPort
}

func (h TcpHeader) GetType() netstack.ProtocolType {
	return netstack.ProtocolTypeTCP
}

func (h TcpHeader) IsFIN() bool {
	return h.BitFlags&TCP_FIN == TCP_FIN
}

func (h *TcpHeader) SetFIN() {
	h.BitFlags |= TCP_FIN
}

func (h TcpHeader) IsSYN() bool {
	return h.BitFlags&TCP_SYN == TCP_SYN
}

func (h *TcpHeader) SetSYN() {
	h.BitFlags |= TCP_SYN
}

func (h TcpHeader) IsRST() bool {
	return h.BitFlags&TCP_RST == TCP_RST
}

func (h *TcpHeader) SetRST() {
	h.BitFlags |= TCP_RST
}

func (h TcpHeader) IsPSH() bool {
	return h.BitFlags&TCP_PSH == TCP_PSH
}

func (h *TcpHeader) SetPSH() {
	h.BitFlags |= TCP_PSH
}

func (h TcpHeader) IsACK() bool {
	return h.BitFlags&TCP_ACK == TCP_ACK
}

func (h *TcpHeader) SetACK() {
	h.BitFlags |= TCP_ACK
}

type TcpOptions struct {
	Mss       uint16
	Nop       uint8
	Nop2      uint8
	WScale    uint8
	SackPerm  uint8
	Sack      uint8
	Timestamp uint8
}

func (o TcpOptions) Marshal() []byte {
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

func (o *TcpOptions) Unmarshal(b []byte) {
	o.Mss = binary.BigEndian.Uint16(b[0:2])
	o.Nop = b[2]
	o.Nop2 = b[3]
	o.WScale = b[4]
	o.SackPerm = b[5]
	o.Sack = b[6]
	o.Timestamp = b[7]
}

// TCP Pseudo Header
type TcpPseudoHeader struct {
	SrcIP    net.IP
	DstIP    net.IP
	Zero     uint8
	Protocol uint8
	Length   uint16
}

func (ph TcpPseudoHeader) Marshal() []byte {
	b := []byte{}
	b = append(b, ph.SrcIP...)
	b = append(b, ph.DstIP...)
	b = append(b, ph.Zero)
	b = append(b, ph.Protocol)
	b = append(b, byte(ph.Length>>8))
	b = append(b, byte(ph.Length))
	return b
}

/*******************************************************************************
	TCP Control Block
*******************************************************************************/

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

func NewTCB() *TCB {
	return &TCB{
		State: TCP_STATE_CLOSED,
	}
}

type SegmentQueue []*netstack.SkBuff

// func (q SegmentQueue) Len() int { return len(q) }
// func (q SegmentQueue) Less(i, j int) bool {
// 	segment1 := q[i].L4Header.(*TcpHeader)
// 	segment2 := q[j].L4Header.(*TcpHeader)
// 	return segment1.SeqNum < segment2.SeqNum
// }

/*******************************************************************************
	TCP Protocol
*******************************************************************************/
type TcpProtocol struct {
	netstack.IProtocol
	ConnTable map[string]*TCB
}

func NewTCP() *TcpProtocol {
	tcp := &TcpProtocol{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeTCP),
		ConnTable: make(map[string]*TCB),
	}
	return tcp
}

/*
	TCP HandleRx algorithm
	- Unmarshal the TCP header
	- Find the TCB. Create new TCB is packet is a SYN.
	- Check the sequence numbers, make sure packet is within the window.
	- Put the packet into the segment processing queue.
*/

func (tcp *TcpProtocol) HandleRx(skb *netstack.SkBuff) {
	// Create a new TCP header
	tcpHeader := &TcpHeader{}

	// Unmarshal the TCP header, handle errors
	if err := tcpHeader.Unmarshal(skb.Data); err != nil {
		skb.Error(err)
		return
	} else {
		skb.StripBytes(tcpHeader.Size())
		skb.L4ProtocolType = netstack.ProtocolTypeTCP
		skb.SrcAddr.Port = tcpHeader.SrcPort
		skb.DestAddr.Port = tcpHeader.DstPort
	}

	// TODO: Handle fragmentation, possibly reassemble

	// Find the TCB for this connection
	connID := ConnectionID(skb.SrcAddr, skb.DestAddr)

	var tcb *TCB
	tcb, ok := tcp.ConnTable[connID]
	if !ok {
		// TCB does not exist. All data is discarded.
		if tcpHeader.IsRST() {
			skb.Error(errors.New("TCP: RST received for non-existent connection"))
			return
		} else {
			tcp.SendEmptyRst(tcpHeader)
		}
	}

	// Handle the TCP packet

	// Check the sequence number, make sure it's in the window
	if tcpHeader.SeqNum < tcb.RecvNXT || tcpHeader.SeqNum > tcb.RecvNXT+tcb.RecvWND {
		skb.Error(errors.New("TCP: SeqNum out of window"))
		return
	}

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
		if tcpHeader.IsSYN() && tcpHeader.IsACK() {
			// Check sequence number

			// Check ACK number

			// Check window size

			// Check urgent pointer

			// Check options

			// Check checksum

			tcb.State = TCP_STATE_ESTABLISHED
		}

	}

	// Update the TCB

	// Everything is good, send to user space
}

func (tcp *TcpProtocol) HandleTx(skb *netstack.SkBuff) {
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

//==============================================================================
// TCP Event Handlers
//==============================================================================

// SendSynAck sends a SYN/ACK packet to the remote TCP in response
// to a SYN packet. The TCB is updated with the remote TCP's ISN.
// The TCBs state is set to TCP_STATE_SYN_RCVD.
func (tcp *TcpProtocol) SendSynAck(skb *netstack.SkBuff, tcb *TCB, requestHeader *TcpHeader) {

	// Create a new TCP header
	newHeader := &TcpHeader{}

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

	newSkb.SrcAddr = skb.DestAddr
	newSkb.DestAddr = skb.SrcAddr
	newSkb.L4ProtocolType = netstack.ProtocolTypeTCP

	// Send to the network layer
	tcp.TxDown(newSkb)

	// Wait for skb to be sent
	newSkb.GetResp()
}

func (tcp *TcpProtocol) SendEmptyRst(tcpHeader *TcpHeader) {
	// Create a new TCP header
	newHeader := &TcpHeader{}

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
func (tcp *TcpProtocol) OpenConnection(sock_meta netstack.SocketMeta) error {
	// Make empty skb
	skb := netstack.NewSkBuff([]byte{})

	if err := set_skb_type(skb); err != nil {
		return err
	}

	// Make TCP header
	header := &TcpHeader{}
	isn := ISN()

	header.SrcPort = sock_meta.SrcAddr.Port
	header.DstPort = sock_meta.DestAddr.Port
	header.BitFlags = TCP_SYN
	header.SeqNum = isn
	header.AckNum = 0
	header.HeaderLen = 5
	header.Window = 0xFFFF
	header.UrgentPtr = 0

	set_tcp_checksum(skb, header)
	skb.PrependBytes(header.Marshal())

	// Send to the network layer
	tcp.TxDown(skb)
	if skbResp := skb.GetResp(); skbResp.Error != nil {
		return skbResp.Error
	}

	// Create a new TCB
	tcb := NewTCB()
	tcb.State = TCP_STATE_SYN_SENT
	tcb.SendUNA = isn
	tcb.SendNXT = isn + 1

	// Update the TCB table
	connID := ConnectionID(sock_meta.SrcAddr, sock_meta.DestAddr)
	tcp.ConnTable[connID] = tcb

	return nil
}

func set_tcp_checksum(skb *netstack.SkBuff, header *TcpHeader) {
	// Set the checksum initially to 0
	header.Checksum = 0

	// Make pseudo header
	// TODO: Handle IPv6
	ph := &TcpPseudoHeader{
		SrcIP:    skb.SrcAddr.IP,
		DstIP:    skb.DestAddr.IP,
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

func (tcp *TcpProtocol) CloseConnection() {
	// This function is meant to be called by the Socket layer,
	// in response to a socket close request.

	// Find the TCB for this connection
	// Update the TCB
	// Send a FIN packet to the remote TCP
	// Wait for an ACK
}

//==============================================================================
// Sequence Number Functions
//==============================================================================

func IsLessThan(isn, seq1, seq2 uint32) bool {
	// Compare using modular arithmetic
	// return (seq1-isn)%0xffffffff < (seq2-isn)%0xffffffff
	return seq1 < seq2
}
