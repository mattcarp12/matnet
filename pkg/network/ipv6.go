package network

import (
	"log"

	"github.com/mattcarp12/go-net/pkg/skbuff"
)

type IPV6 struct {
	skbuff.SkBuffReaderWriter
}

func NewIPV6() IPV6 {
	return IPV6{
		SkBuffReaderWriter: skbuff.NewSkBuffChannels(),
	}
}

func (ipv6 IPV6) HandleRX(skb *skbuff.Sk_buff) error {
	log.Printf("IPV6: %v", skb)
	return nil
}

func (ipv6 IPV6) HandleTX(skb *skbuff.Sk_buff) error {
	log.Printf("IPV6: %v", skb)
	return nil
}
