package link

import (
	"github.com/mattcarp12/go-net/pkg/headers"
	"github.com/mattcarp12/go-net/pkg/protocols"
	"github.com/mattcarp12/go-net/pkg/skbuff"
)

func eth_rx_handle(skb *skbuff.Sk_buff) error {
	// Parse the ethernet header
	skb.Ethernet = &headers.EthernetHeader{}

	err := skb.Ethernet.Unmarshal(skb.Data)
	if err != nil {
		return err
	}

	// Fill out L3 protocol field in the skbuff
	skb.L3_Protocol = protocols.L3_Protocol_From_EtherType(skb.Ethernet.EtherType)

	return nil
}
