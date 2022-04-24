package network

import (
	"log"

	"github.com/mattcarp12/go-net/pkg/skbuff"
)

type ARP struct {
	skbuff.SkBuffReaderWriter
}

func NewARP() ARP {
	return ARP{
		SkBuffReaderWriter: skbuff.NewSkBuffChannels(),
	}
}

func (arp ARP) HandleRX(skb *skbuff.Sk_buff) error {
	log.Printf("ARP: %v", skb)
	return nil
}

func (arp ARP) HandleTX(skb *skbuff.Sk_buff) error {
	log.Printf("ARP: %v", skb)
	return nil
}
