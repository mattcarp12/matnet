package netstack

import (
	"log"
	"os"
)

var skb_log = log.New(os.Stdout, "[SKB] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

/*
	SkBuff is the struct that represents a packet as it moves through the stack.
*/
type SkBuff struct {
	Data []byte

	ProtocolType

	iface NetworkInterface

	l2_header L2Header

	l3_header L3Header

	l4_header L4Header

	socket Socket

	route route

	resp chan SkbResponse
}

func (s *SkBuff) GetBytes() []byte {
	return s.Data
}

func NewSkBuff(data []byte) *SkBuff {
	return &SkBuff{
		Data: data,
		resp: make(chan SkbResponse),
	}
}

func (skb *SkBuff) GetType() ProtocolType {
	return skb.ProtocolType
}

func (skb *SkBuff) SetType(t ProtocolType) {
	skb.ProtocolType = t
}

func (skb *SkBuff) PrependBytes(b []byte) {
	skb.Data = append(b, skb.Data...)
}

func (skb *SkBuff) StripBytes(n int) {
	skb.Data = skb.Data[n:]
}

func (skb *SkBuff) GetRoute() route {
	return skb.route
}

func (skb *SkBuff) SetRoute(r route) {
	skb.route = r
}

func (skb *SkBuff) GetNetworkInterface() NetworkInterface {
	return skb.iface
}

func (skb *SkBuff) SetNetworkInterface(iface NetworkInterface) {
	skb.iface = iface
}

func (skb *SkBuff) Resp() SkbResponse {
	skb_log.Printf("Called Resp")
	resp := <-skb.resp
	skb_log.Printf("Resp: %+v", resp)
	return resp
}

/*
	Header Get/Set methods

	These are used by the protocol handlers to set the L2/L3/L4 headers
	of the skbuff.
*/

func (skb *SkBuff) GetL2Header() L2Header {
	return skb.l2_header
}

func (skb *SkBuff) SetL2Header(h L2Header) {
	skb.l2_header = h
}

func (skb *SkBuff) GetL3Header() L3Header {
	return skb.l3_header
}

func (skb *SkBuff) SetL3Header(h L3Header) {
	skb.l3_header = h
}

func (skb *SkBuff) GetL4Header() L4Header {
	return skb.l4_header
}

func (skb *SkBuff) SetL4Header(h L4Header) {
	skb.l4_header = h
}

/*
	Socket methods
*/
func (skb *SkBuff) GetSocket() Socket {
	return skb.socket
}

func (skb *SkBuff) SetSocket(s Socket) {
	skb.socket = s
}

/*
	SkBuffChannels are used to pass packets up and down the stack.
*/

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

/*
	SkbResponse is used to return a response from the network stack about the
	status of a packet.
*/

type SkbResponse struct {
	err           error
	status        int
	bytes_read    int
	bytes_written int
}

func (r SkbResponse) Error() error {
	return r.err
}

func (r SkbResponse) Status() int {
	return r.status
}

func (r SkbResponse) BytesRead() int {
	return r.bytes_read
}

func (r SkbResponse) BytesWritten() int {
	return r.bytes_written
}

func SkbErrorResp(err error) SkbResponse {
	return SkbResponse{
		err: err,
	}
}

func SkbWriteResp(bytes_written int) SkbResponse {
	return SkbResponse{
		bytes_written: bytes_written,
	}
}

func (skb *SkBuff) Error(err error) {
	skb.resp <- SkbErrorResp(err)
}

func (skb *SkBuff) TxSuccess(bytes_written int) {
	skb_log.Printf("Called TxSuccess with %d bytes written", bytes_written)
	skb.resp <- SkbWriteResp(bytes_written)
}
