package linklayer

import (
	"log"

	"github.com/mattcarp12/go-net/pkg/entities/ethernet"
	"github.com/mattcarp12/go-net/pkg/entities/protocols"
)

type Ethernet struct {
	protocols.PDUReaderWriter
	layer protocols.Layer
}

func NewEthernet() *Ethernet {
	return &Ethernet{
		PDUReaderWriter: protocols.NewPDUChannels(),
	}
}

func (eth *Ethernet) GetType() protocols.ProtocolType {
	return protocols.ProtocolTypeEthernet
}

func (eth *Ethernet) HandleRx(pdu protocols.PDU) {
	// Parse the ethernet header
	eh, err := ethernet.ParseEthernetHeader(pdu.GetBytes())
	if err != nil {
		log.Printf("Error parsing ethernet header: %v", err)
	}

	// Set skb type to the next layer type
	nextLayerType, err := eh.GetNextLayerType()
	if err != nil {
		log.Printf("Error getting next layer type: %v", err)
	}
	pdu.SetType(nextLayerType)

	// Strip ethernet header
	eh.StripFromPDU(pdu)

	// Pass to next layer
	eth.layer.GetNextLayer().RxChan() <- pdu
}

func (eth *Ethernet) HandleTx(pdu protocols.PDU) {
	// TODO: Implement
}
