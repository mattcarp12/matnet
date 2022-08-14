package transportlayer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/mattcarp12/matnet/netstack/util"

	"github.com/mattcarp12/matnet/netstack"
)

// =============================================================================
// TCP Header
// =============================================================================

var ErrInvalidTCPHeader = errors.New("invalid TCP header")

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
		case TCPOptionKindMSS:
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
	TCP   *TCPProtocol
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

	RxChan       chan TCPBuffer
	RxChanSorted chan TCPBuffer
	RxQueue      *util.Heap[TCPBuffer]
	QuitChan     chan struct{}

	// Helper metadata for transmitting SkBuffs
	SrcAddr netstack.SockAddr
	DstAddr netstack.SockAddr
	TxIface netstack.NetworkInterface

	Log *log.Logger
}

const (
	TCP_QUEUE_SIZE = 1024
)

func (tcp *TCPProtocol) NewTCB(connID string) *TCB {

	rxQueue := util.NewHeap(tcpBuffLess)

	tcb := &TCB{
		TCP:          tcp,
		ID:           connID,
		State:        TCP_STATE_CLOSED,
		RxChan:       make(chan TCPBuffer, TCP_QUEUE_SIZE),
		RxChanSorted: make(chan TCPBuffer, TCP_QUEUE_SIZE),
		QuitChan:     make(chan struct{}),
		RxQueue:      rxQueue,
		RecvWND:      0xffff,
		Log:          tcp.Log,
	}

	tcp.ConnTable[connID] = tcb

	go tcb.MainLoop()

	return tcb
}

type TCPBuffer struct {
	Header *TCPHeader
	SkBuff *netstack.SkBuff
}

func tcpBuffLess(a, b TCPBuffer) bool {
	return a.Header.SeqNum < b.Header.SeqNum
}

/*
Each TCP connection is handled in it's own goroutine.
This is the main loop of the connection.
*/
func (tcb *TCB) MainLoop() {
	for {
		select {
		// RxChan is where we receive packets from the network stack.
		// They are not in sorted order, so we need to sort them.
		case skb := <-tcb.RxChan:
			tcb.sortSegment(skb)

		// RxChanSorted are the packets we have received in sorted order,
		// starting with RecvNXT sequence number.
		case skb := <-tcb.RxChanSorted:
			tcb.handleSegmentArrives(skb)

		// QuitChan is where we receive a signal to quit (obvi).
		case <-tcb.QuitChan:
			return
		}
	}
}

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
	ErrConnectionNoExist     = errors.New("connection does not exist")
	ErrConnectionNoExistRST  = errors.New("received RST for non-existent connection")
	ErrConnectionIllegal     = errors.New("illegal connection")
	ErrConnectionClosing     = errors.New("connection is closing")
	ErrInvalidState          = errors.New("tcp: invalid state")
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
	- Find the TCB. Error if TCB does not exist.
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
	if !ok || tcb == nil {
		// TCB does not exist. All data is discarded.
		if tcpHeader.IsRST() {
			skb.Error(ErrConnectionNoExistRST)
			return
		}

		tcp.SendEmptyRst(tcpHeader)
	}

	tcp.Log.Printf("\n\n********************************************************************\nRECEIVED TCP SEGMENT\n")
	tcp.Log.Printf("TCP Header: %+v\n", tcpHeader)
	tcp.Log.Printf("TCB: %+v\n", tcb)
	tcp.Log.Printf("Header IsSyn: %v\n", tcpHeader.IsSYN())
	tcp.Log.Printf("Header IsACK: %v\n", tcpHeader.IsACK())
	tcp.Log.Printf("Header IsFIN: %v\n", tcpHeader.IsFIN())
	tcp.Log.Printf("\n********************************************************************\n\n")

	// Put the packet into the TCB's RxQueue
	tcb.RxChan <- TCPBuffer{
		Header: tcpHeader,
		SkBuff: skb,
	}
}

// This is where incoming packets are checked for seq numbers, and
// put into the TCB's processing queue in the correct order.
func (tcb *TCB) sortSegment(tcpBuff TCPBuffer) {
	header := tcpBuff.Header
	skb := tcpBuff.SkBuff

	// If we're in the SYN-SENT state, the usual processing does not apply.
	// We must handle this case specially.
	if tcb.State == TCP_STATE_SYN_SENT {
		if err := tcb.HandleSynSent(tcpBuff); err != nil {
			skb.Error(err)
		}
		return
	}

	// First check the sequence number, make sure it's in the window
	if header.SeqNum < tcb.RecvNXT || header.SeqNum > tcb.RecvNXT+tcb.RecvWND {
		tcb.Log.Printf("TCP: SeqNum out of window\n")
		skb.Error(ErrInvalidSequenceNumber)
		return
	}

	// The new segment is within the window, so put it in the TCB's RxQueue.
	tcb.RxQueue.Push(tcpBuff)

	// Check the next packet in the queue, see if it's ready to be processed
	tcpBuff2 := tcb.RxQueue.Peek()

	if tcpBuff2.Header.SeqNum != tcb.RecvNXT {
		return
	}

	// If we're here, we have a packet that matches the next sequence number.
	// So enqueue it to the sorted channel, and also all the next packets
	// that are ready to be processed, incrementing RecvNXT as we go.
	for {
		if tcb.RxQueue.Len() == 0 {
			break
		}

		tcpBuff := tcb.RxQueue.Peek()

		// Check the sequence number, make sure it equals RecvNXT
		if tcpBuff.Header.SeqNum != tcb.RecvNXT {
			break
		}

		// If we're here, we have a packet that matches the next sequence number.
		tcpBuff = tcb.RxQueue.Pop()
		tcb.RxChanSorted <- tcpBuff

		// At this point, the TCP Header has been stripped from the skbuff data buffer,
		// and all that is left is the application data. So we can increment the RecvNXT,
		// by the length of the application data.
		tcb.RecvNXT += uint32(len(tcpBuff.SkBuff.Data))
	}
}

// This is where the main TCP logic happens. This function should be called with packets
// in sequence number order.
func (tcb *TCB) handleSegmentArrives(tcpBuff TCPBuffer) {
	tcb.Log.Printf("\n\nTCP: handleSegmentArrives\n\n")

	header := tcpBuff.Header
	skb := tcpBuff.SkBuff

	// Handle the TCP packet

	// If we received an ACK, set SendNXT to the AckNum
	if header.IsACK() {
		tcb.SendNXT = header.AckNum
	}

	switch tcb.State {
	default:
		skb.Error(ErrInvalidState)
	case TCP_STATE_LISTEN:
		if header.IsSYN() {
			tcb.SendSynAck(tcpBuff)
			return
		}
	case TCP_STATE_SYN_RCVD:
		// Here we sent a SYN+ACK, and now we are waiting for an ACK.
		if header.IsACK() && header.AckNum == tcb.SendNXT {
			tcb.State = TCP_STATE_ESTABLISHED
		}
	case TCP_STATE_SYN_SENT:
		// This should have already been handled in sortSegment.
		tcb.Log.Printf("Error: Should not be in SYN_SENT state at this point.\n")
		skb.Error(ErrInvalidState)

	case TCP_STATE_FIN_WAIT_1:
		// Here we sent a FIN segment, and now we are waiting for a FIN+ACK.
		if header.IsACK() && header.AckNum == tcb.SendNXT {
			tcb.State = TCP_STATE_FIN_WAIT_2

			// If we received a FIN+ACK, we need to ACK it.
			if header.IsFIN() {
				tcb.Log.Printf("\n\nRECEIVED A FIN+ACK\n\n")
				tcb.SendAck(tcpBuff)
			}
		}

	case TCP_STATE_FIN_WAIT_2:
		// Here we received an ACK for our FIN segment, and now we are waiting for a FIN+ACK.
		if header.IsACK() && header.IsFIN() && header.AckNum == tcb.SendNXT {
			// Need to ACK the FIN
			tcb.SendAck(tcpBuff)
			tcb.State = TCP_STATE_CLOSE_WAIT
		}

	case TCP_STATE_ESTABLISHED:

	}
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
func (tcb *TCB) SendSynAck(tcpBuff TCPBuffer) {
	requestHeader := tcpBuff.Header
	skb := tcpBuff.SkBuff

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
	tcb.TCP.TxDown(newSkb)

	// Wait for skb to be sent
	newSkb.GetResp()
}

func (tcb *TCB) SendAck(tcpBuff TCPBuffer) {
	requestHeader := tcpBuff.Header
	skb := tcpBuff.SkBuff

	tcb.Log.Printf("Sending Ack for %+v\n", requestHeader)
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
		tcb.Log.Printf("Error getting rxIface: %v\n", err)
		return
	}

	newSkb.SetTxIface(rxIface)

	setSkbType(newSkb)

	tcb.TCP.TxDown(newSkb)

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
	tcb.SrcAddr = srcAddr
	tcb.DstAddr = dstAddr
	tcb.TxIface = iface

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

// CloseConnection is meant to be called by the Socket layer,
// in response to a socket close request.
func (tcp *TCPProtocol) CloseConnection(srcAddr, dstAddr netstack.SockAddr) error {
	tcp.Log.Printf("CloseConnection: %v -> %v\n", srcAddr, dstAddr)

	// Get the TCB
	connID := ConnectionID(srcAddr, dstAddr)

	tcb, ok := tcp.ConnTable[connID]
	if !ok {
		tcp.Log.Printf("CloseConnection: No TCB for %v\n", connID)
		return fmt.Errorf("CloseConnection: No TCB for %v. %w", connID, ErrConnectionNoExist)
	}

	switch tcb.State {
	case TCP_STATE_CLOSED:
		tcp.Log.Printf("CloseConnection: Connection %v already closed\n", connID)
		return fmt.Errorf("CloseConnection: Connection %v already closed. %w", connID, ErrConnectionIllegal)

	case TCP_STATE_LISTEN, TCP_STATE_SYN_SENT:
		// Delete the TCB
		delete(tcp.ConnTable, connID)
		tcp.Log.Printf("CloseConnection: Connection %v closed\n", connID)
		return nil

	// Normal case
	case TCP_STATE_SYN_RCVD, TCP_STATE_ESTABLISHED:
		// Queue a FIN, enter FIN_WAIT_1 state
		tcb.State = TCP_STATE_FIN_WAIT_1
		return tcp.SendFin(tcb)

	case TCP_STATE_CLOSE_WAIT:
		// Send a FIN, enter CLOSING state
		tcb.State = TCP_STATE_CLOSING
		return tcp.SendFin(tcb)

	case TCP_STATE_FIN_WAIT_1, TCP_STATE_FIN_WAIT_2, TCP_STATE_CLOSING, TCP_STATE_LAST_ACK, TCP_STATE_TIME_WAIT:
		return fmt.Errorf("CloseConnection: %w", ErrConnectionClosing)
	}

	return nil
}

// SendFin sends a FIN packet to the remote TCP
func (tcp *TCPProtocol) SendFin(tcb *TCB) error {
	// Make empty skb
	skb := netstack.NewSkBuff([]byte{})

	skb.SetSrcAddr(tcb.SrcAddr)
	skb.SetDstAddr(tcb.DstAddr)
	skb.SetTxIface(tcb.TxIface)

	if err := setSkbType(skb); err != nil {
		return err
	}

	// Make TCP header
	header := &TCPHeader{}

	header.SrcPort = tcb.SrcAddr.Port
	header.DstPort = tcb.DstAddr.Port
	header.BitFlags = TCP_FIN | TCP_ACK
	header.SeqNum = tcb.SendNXT
	header.AckNum = tcb.RecvNXT
	header.HeaderLen = 5
	header.Window = 0xFFFF
	header.UrgentPtr = 0

	setTCPChecksum(skb, header)
	skb.SetL4Header(header)
	skb.PrependBytes(header.Marshal())

	// Send to the network layer
	tcp.TxDown(skb)

	if skbResp := skb.GetResp(); skbResp.Error != nil {
		tcp.Log.Printf("SendFin: Error sending FIN: %v\n", skbResp.Error)
		return skbResp.Error
	}

	return nil
}

// HandleSynSent is called when a packet is received and the state is TCP_STATE_SYN_SENT
func (tcb *TCB) HandleSynSent(tcpBuff TCPBuffer) error {
	header := tcpBuff.Header

	tcb.Log.Printf("HandleSynSent: %v\n", header)
	// First check the ACK bit
	if header.IsACK() {
		if header.AckNum <= tcb.SendISN || header.AckNum > tcb.SendNXT {
			tcb.TCP.SendEmptyRst(header)
			return fmt.Errorf("HandleSynSent: %w", ErrInvalidSequenceNumber)
		}
	}

	// Check if the ACK number is valid
	if header.AckNum < tcb.SendUNA && header.AckNum > tcb.SendNXT {
		return fmt.Errorf("HandleSynSent: %w", ErrInvalidAckNumber)
	}

	// Second check the RST bit
	if header.IsRST() {
		// TODO: How to signal to the user that the connection was reset?

		// Remove the TCB
		delete(tcb.TCP.ConnTable, tcb.ID)

		return ErrConnectionReset
	}

	// TODO: Check Security and Precedence bits

	// Fourth check the SYN bit
	if header.IsSYN() {
		tcb.RecvNXT = header.SeqNum + 1
		tcb.RecvISN = header.SeqNum
		tcb.SendUNA = header.AckNum

		// Update the state
		if tcb.SendUNA > tcb.SendISN {
			tcb.State = TCP_STATE_ESTABLISHED
			tcb.SendAck(tcpBuff)
		} else {
			tcb.State = TCP_STATE_SYN_RCVD
			tcb.SendSynAck(tcpBuff)
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
