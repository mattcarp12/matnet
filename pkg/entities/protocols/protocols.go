package protocols

type Protocol interface {
	GetType() ProtocolType
	PDUReaderWriter
	HandleRx(PDU)
	HandleTx(PDU)
}

func RxLoop(protocol Protocol) {
	for {
		// Network protocol reads pdu from it's rx_chan
		skb := <-protocol.RxChan()

		// Handle sk_buff
		protocol.HandleRx(skb)
	}
}

func TxLoop(protocol Protocol) {
	for {
		// Network protocol reads pdu from it's tx_chan
		skb := <-protocol.TxChan()

		// Handle sk_buff
		protocol.HandleTx(skb)
	}
}
