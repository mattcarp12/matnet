package network

import (
	"log"

	"github.com/mattcarp12/go-net/pkg/skbuff"
)

type IPV4 struct {
	skbuff.SkBuffReaderWriter
}

func NewIPV4() IPV4 {
	return IPV4{
		SkBuffReaderWriter: skbuff.NewSkBuffChannels(),
	}
}

func (ipv4 IPV4) HandleRX(skb *skbuff.Sk_buff) error {
	log.Printf("IPV4: %v", skb)
	return nil
}

func (ipv4 IPV4) HandleTX(skb *skbuff.Sk_buff) error {
	log.Printf("IPV4: %v", skb)
	return nil
}