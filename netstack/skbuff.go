package netstack

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
}

func (s *SkBuff) GetBytes() []byte {
	return s.Data
}

func NewSkBuff(data []byte) *SkBuff {
	return &SkBuff{
		Data: data,
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

func (skb *SkBuff) GetNetworkInterface() NetworkInterface {
	return skb.iface
}

func (skb *SkBuff) SetNetworkInterface(iface NetworkInterface) {
	skb.iface = iface
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
