package netstack

import (
	"errors"
	"log"
	"net"
	"os"
	"strconv"
)

var skb_log = log.New(os.Stdout, "[SKB] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

//=============================================================================
// SockAddr -- generic structure for network addresses
// Can hold either an IPv4 or IPv6 address
//=============================================================================
type SockAddr struct {
	Port uint16
	IP   net.IP
}

type AddressType uint

const (
	AddressTypeIPv4 = iota
	AddressTypeIPv6
	AddressTypeUnknown
)

var ErrInvalidAddressType = errors.New("invalid address type")

func (s SockAddr) String() string {
	return s.IP.String() + ":" + strconv.Itoa(int(s.Port))
}

func (s SockAddr) GetType() AddressType {
	if s.IP.To4() != nil {
		return AddressTypeIPv4
	} else if s.IP.To16() != nil {
		return AddressTypeIPv6
	} else {
		return AddressTypeUnknown
	}
}

//=============================================================================
// SkBuff is the struct that represents a packet as it moves
// through the networking stack.
//=============================================================================
type SkBuff struct {
	Data         []byte
	protocolType ProtocolType
	rxIface      NetworkInterface
	txIface      NetworkInterface
	srcAddr      SockAddr
	dstAddr      SockAddr
	l2Header     L2Header
	l3Header     L3Header
	l4Header     L4Header
	RespChan     chan SkbResponse

	// Resp SkbResponse
}

// NewSkbuff creates a new SkBuff
func NewSkBuff(data []byte) *SkBuff {
	return &SkBuff{
		Data:         data,
		protocolType: ProtocolTypeUnknown,
		RespChan:     make(chan SkbResponse),
	}
}

// PrependBytes is used to prepend the data payload with
// protocol headers.
func (skb *SkBuff) PrependBytes(b []byte) {
	skb.Data = append(b, skb.Data...)
}

// StripBytes is used to remove the protocol headers from the
// data payload.
func (skb *SkBuff) StripBytes(n int) {
	skb.Data = skb.Data[n:]
}

func (skb *SkBuff) GetResp() SkbResponse {
	resp := <-skb.RespChan
	return resp
}

func (skb *SkBuff) GetType() ProtocolType {
	return skb.protocolType
}

func (skb *SkBuff) SetType(protocolType ProtocolType) {
	skb.protocolType = protocolType
}

func (skb *SkBuff) GetRxIface() (NetworkInterface, error) {
	if skb.rxIface == nil {
		return nil, errors.New("network interface not set")
	}
	return skb.rxIface, nil
}

func (skb *SkBuff) SetRxIface(netdev NetworkInterface) {
	skb.rxIface = netdev
}

func (skb *SkBuff) GetTxIface() (NetworkInterface, error) {
	if skb.txIface == nil {
		return nil, errors.New("network interface not set")
	}
	return skb.txIface, nil
}

func (skb *SkBuff) SetTxIface(netdev NetworkInterface) {
	skb.txIface = netdev
}

func (skb *SkBuff) GetL2Header() (L2Header, error) {
	if skb.l2Header == nil {
		return nil, errors.New("L2 header not set")
	}
	return skb.l2Header, nil
}

func (skb *SkBuff) SetL2Header(header L2Header) {
	skb.l2Header = header
}

func (skb *SkBuff) GetL3Header() (L3Header, error) {
	if skb.l3Header == nil {
		return nil, errors.New("L3 header not set")
	}
	return skb.l3Header, nil
}

func (skb *SkBuff) SetL3Header(header L3Header) {
	skb.l3Header = header
}

func (skb *SkBuff) GetL4Header() (L4Header, error) {
	if skb.l4Header == nil {
		return nil, errors.New("L4 header not set")
	}
	return skb.l4Header, nil
}

func (skb *SkBuff) SetL4Header(header L4Header) {
	skb.l4Header = header
}

func (skb *SkBuff) GetSrcAddr() SockAddr {
	return skb.srcAddr
}

func (skb *SkBuff) SetSrcAddr(addr SockAddr) {
	skb.srcAddr = addr
}

func (skb *SkBuff) GetDstAddr() SockAddr {
	return skb.dstAddr
}

func (skb *SkBuff) SetDstAddr(addr SockAddr) {
	skb.dstAddr = addr
}

func (skb *SkBuff) GetSrcIP() net.IP {
	return skb.srcAddr.IP
}

func (skb *SkBuff) SetSrcIP(ip net.IP) {
	skb.srcAddr.IP = ip
}

func (skb *SkBuff) GetSrcPort() uint16 {
	return skb.srcAddr.Port
}

func (skb *SkBuff) SetSrcPort(port uint16) {
	skb.srcAddr.Port = port
}

func (skb *SkBuff) GetDstIP() net.IP {
	return skb.dstAddr.IP
}

func (skb *SkBuff) SetDstIP(ip net.IP) {
	skb.dstAddr.IP = ip
}

func (skb *SkBuff) GetDstPort() uint16 {
	return skb.dstAddr.Port
}

func (skb *SkBuff) SetDstPort(port uint16) {
	skb.dstAddr.Port = port
}

//=============================================================================
// SkBuffChannels are used to pass packets up and down the stack.
//=============================================================================

type SkBuffReader interface {
	RxChan() chan *SkBuff
}

type SkBuffWriter interface {
	TxChan() chan *SkBuff
}

type SkBuffReaderWriter interface {
	SkBuffReader
	SkBuffWriter
}

type SkBuffChannels struct {
	rx_chan chan *SkBuff
	tx_chan chan *SkBuff
}

func NewSkBuffChannels() SkBuffChannels {
	return SkBuffChannels{
		rx_chan: make(chan *SkBuff),
		tx_chan: make(chan *SkBuff),
	}
}

func (skb_channels SkBuffChannels) RxChan() chan *SkBuff {
	return skb_channels.rx_chan
}

func (skb_channels SkBuffChannels) TxChan() chan *SkBuff {
	return skb_channels.tx_chan
}

//==============================================================================
// SkbResponse is used to return a response from the network stack about the
// status of a packet.
//==============================================================================

type SkbResponse struct {
	Error        error
	Status       int
	BytesRead    int
	BytesWritten int
}

func SkbErrorResp(err error) SkbResponse {
	return SkbResponse{
		Error: err,
	}
}

func SkbWriteResp(bytes_written int) SkbResponse {
	return SkbResponse{
		BytesWritten: bytes_written,
	}
}

func (skb *SkBuff) Error(err error) {
	skb.RespChan <- SkbErrorResp(err)
}

func (skb *SkBuff) TxSuccess() {
	skb.RespChan <- SkbWriteResp(len(skb.Data))
}
