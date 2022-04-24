package transport

import "github.com/mattcarp12/go-net/pkg/skbuff"

type TransportLayer struct {
	rx_chan chan *skbuff.Sk_buff
	tx_chan chan *skbuff.Sk_buff

	// tcp  TCP
	// udp  UDP
}
