package netstack

import (
	"errors"
	"log"
	"os"
)

var skb_log = log.New(os.Stdout, "[SKB] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

/*
	SkBuff is the struct that represents a packet as it moves through the stack.
*/
type SkBuff struct {
	Data             []byte
	ProtocolType     ProtocolType
	NetworkInterface NetworkInterface
	L2Header         L2Header
	L3Header         L3Header
	L4Header         L4Header
	RespChan         chan SkbResponse

	// Resp SkbResponse
}

func NewSkBuff(data []byte) *SkBuff {
	return &SkBuff{
		Data:         data,
		ProtocolType: ProtocolTypeUnknown,
		RespChan:     make(chan SkbResponse),
	}
}

func (skb *SkBuff) PrependBytes(b []byte) {
	skb.Data = append(b, skb.Data...)
}

func (skb *SkBuff) StripBytes(n int) {
	skb.Data = skb.Data[n:]
}

func (skb *SkBuff) GetResp() SkbResponse {
	resp := <-skb.RespChan
	return resp
}

func (skb *SkBuff) GetType() ProtocolType {
	return skb.ProtocolType
}

func (skb *SkBuff) SetType(protocolType ProtocolType) {
	skb.ProtocolType = protocolType
}

func (skb *SkBuff) GetNetworkInterface() (NetworkInterface, error) {
	if skb.NetworkInterface == nil {
		return nil, errors.New("network interface not set")
	}
	return skb.NetworkInterface, nil
}

func (skb *SkBuff) SetNetworkInterface(netdev NetworkInterface) {
	skb.NetworkInterface = netdev
}

func (skb *SkBuff) GetL2Header() L2Header {
	return skb.L2Header
}

func (skb *SkBuff) SetL2Header(header L2Header) {
	skb.L2Header = header
}

func (skb *SkBuff) GetL3Header() L3Header {
	return skb.L3Header
}

func (skb *SkBuff) SetL3Header(header L3Header) {
	skb.L3Header = header
}

func (skb *SkBuff) GetL4Header() L4Header {
	return skb.L4Header
}

func (skb *SkBuff) SetL4Header(header L4Header) {
	skb.L4Header = header
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
