package ipv6

import (
	"log"

	"github.com/mattcarp12/go-net/netstack"
)

type IPV6 struct {
	netstack.IProtocol
}

func NewIPV6() *IPV6 {
	return &IPV6{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeIPv6),
	}
}

func (ipv6 *IPV6) HandleRx(skb *netstack.SkBuff) {
	log.Printf("IPV6: %v", skb)
}

func (ipv6 *IPV6) HandleTx(skb *netstack.SkBuff) {
	log.Printf("IPV6: %v", skb)
}
