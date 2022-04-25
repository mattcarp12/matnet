package networklayer

import (
	"github.com/mattcarp12/go-net/pkg/entities/protocols"
	"log"
)

type IPV6 struct {
	protocols.PDUReaderWriter
}

func NewIPV6() *IPV6 {
	return &IPV6{
		PDUReaderWriter: protocols.NewPDUChannels(),
	}
}

func (ipv6 *IPV6) HandleRx(pdu protocols.PDU) {
	log.Printf("IPV6: %v", pdu)
}

func (ipv6 *IPV6) HandleTx(pdu protocols.PDU) {
	log.Printf("IPV6: %v", pdu)
}

func (ipv6 *IPV6) GetType() protocols.ProtocolType {
	return protocols.ProtocolTypeIPv6
}
