package networklayer

import (
	"log"
	"github.com/mattcarp12/go-net/pkg/entities/protocols"
)

type ARP struct {
	protocols.PDUReaderWriter
	layer protocols.Layer
}

func NewARP() *ARP {
	return &ARP{
		PDUReaderWriter: protocols.NewPDUChannels(),
	}
}

func (arp *ARP) GetType() protocols.ProtocolType {
	return protocols.ProtocolTypeARP
}

func (arp *ARP) HandleRx(pdu protocols.PDU) {
	log.Printf("ARP: %v", pdu)
}

func (arp *ARP) HandleTx(pdu protocols.PDU) {
	log.Printf("ARP: %v", pdu)
}
