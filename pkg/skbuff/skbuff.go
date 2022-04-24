package skbuff

import (
	"github.com/mattcarp12/go-net/pkg/headers"
	"github.com/mattcarp12/go-net/pkg/protocols"
)

type Sk_buff struct {
	// Raw data from the wire
	Data []byte

	// Ethernet Header
	Ethernet *headers.EthernetHeader

	// Network Layer Protocol
	L3_Protocol protocols.NetworkProtocolID

	// Transport Layer Protocol
	L4_Protocol protocols.NetworkProtocolID
}

func New(data []byte) *Sk_buff {
	return &Sk_buff{
		Data: data,
	}
}

type SkBuffReaderWriter interface {
	RxChan() chan *Sk_buff
	TxChan() chan *Sk_buff
}

type SkBuffChannels struct {
	rx_chan chan *Sk_buff
	tx_chan chan *Sk_buff
}

func NewSkBuffChannels() SkBuffChannels {
	return SkBuffChannels{
		rx_chan: make(chan *Sk_buff),
		tx_chan: make(chan *Sk_buff),
	}
}

func (s SkBuffChannels) RxChan() chan *Sk_buff {
	return s.rx_chan
}

func (s SkBuffChannels) TxChan() chan *Sk_buff {
	return s.tx_chan
}
