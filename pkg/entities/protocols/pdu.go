package protocols

type PDU interface {
	GetType() ProtocolType
	SetType(ProtocolType)
	GetBytes() []byte
	AppendBytes([]byte)
	StripBytes(n int)
}

type IPDU struct {
	Data []byte
	Type ProtocolType
}

func (pdu *IPDU) GetType() ProtocolType {
	return pdu.Type
}

func (pdu *IPDU) SetType(t ProtocolType) {
	pdu.Type = t
}

func (pdu *IPDU) AppendBytes(b []byte) {
	pdu.Data = append(pdu.Data, b...)
}

func (pdu *IPDU) StripBytes(n int) {
	pdu.Data = pdu.Data[n:]
}

type PDUReader interface {
	RxChan() chan PDU
}

type PDUWriter interface {
	TxChan() chan PDU
}

type PDUReaderWriter interface {
	PDUReader
	PDUWriter
}

type PDUChannels struct {
	rx_chan chan PDU
	tx_chan chan PDU
}

func NewPDUChannels() PDUChannels {
	return PDUChannels{
		rx_chan: make(chan PDU),
		tx_chan: make(chan PDU),
	}
}

func (pdu_channels PDUChannels) RxChan() chan PDU {
	return pdu_channels.rx_chan
}

func (pdu_channels PDUChannels) TxChan() chan PDU {
	return pdu_channels.tx_chan
}
