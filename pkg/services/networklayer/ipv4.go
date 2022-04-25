package networklayer

import (
	"log"
	"github.com/mattcarp12/go-net/pkg/entities/protocols"
)

type IPV4 struct {
	protocols.PDUReaderWriter
	layer protocols.Layer
}

var _ protocols.Protocol = &IPV4{}

func NewIPV4() *IPV4 {
	return &IPV4{
		PDUReaderWriter: protocols.NewPDUChannels(),
	}
}

func (ipv4 *IPV4) HandleRx(pdu protocols.PDU) {
	log.Printf("IPV4: %v", pdu)
}

func (ipv4 *IPV4) HandleTx(pdu protocols.PDU) {
	log.Printf("IPV4: %v", pdu)
}

func (ipv4 *IPV4) GetType() protocols.ProtocolType {
	return protocols.ProtocolTypeIPv4
}
