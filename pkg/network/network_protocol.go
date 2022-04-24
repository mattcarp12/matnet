package network

import "github.com/mattcarp12/go-net/pkg/skbuff"

const (
	Protocol_IPV4 uint8 = iota
	Protocol_IPV6
	Protocol_ARP
)

type NetworkProtocol interface {
	skbuff.SkBuffReaderWriter
	HandleRX(*skbuff.Sk_buff) error
	HandleTX(*skbuff.Sk_buff) error
}
