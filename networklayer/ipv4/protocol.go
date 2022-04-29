package ipv4

import (
	"log"

	"github.com/mattcarp12/go-net/netstack"
)

type IPV4 struct {
	netstack.IProtocol
}

func NewIPV4() *IPV4 {
	return &IPV4{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeIPv4),
	}
}

func (ipv4 *IPV4) HandleRx(skb *netstack.SkBuff) {
	log.Printf("IPV4: %v", skb)
}

func (ipv4 *IPV4) HandleTx(skb *netstack.SkBuff) {
	log.Printf("IPV4: %v", skb)
}
