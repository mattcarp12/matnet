package link

import (
	"log"
	"net"

	"github.com/mattcarp12/go-net/pkg/skbuff"
	"github.com/mattcarp12/go-net/pkg/tuntap"
)

type LinkLayer struct {
	skbuff.SkBuffReaderWriter

	// TODO: support multiple devices
	tap  TAPDevice
	loop LoopbackDevice
}

func (ll LinkLayer) RxLoop(dev NetworkDevice) {
	for {
		frame := make([]byte, 1500)
		skb := skbuff.New(frame) // TODO: use skb pool

		if err := dev.Read(skb); err != nil {
			log.Printf("Error reading from device: %v", err)
			continue
		}

		if err := dev.Handle(skb); err != nil {
			log.Printf("Error handling skb: %v", err)
			continue
		}

		ll.RxChan() <- skb
	}
}

func NewLinkLayer(tap TAPDevice, loop LoopbackDevice) LinkLayer {
	ll := LinkLayer{}
	ll.SkBuffReaderWriter = skbuff.NewSkBuffChannels()
	ll.tap = tap
	ll.loop = loop
	return ll
}

func Init() LinkLayer {
	tap := NewTap(tuntap.TapInit("tap0", tuntap.DefaultIPv4Addr),
		net.IPv4(10, 88, 45, 69), net.HardwareAddr{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad})
	loop := NewLoopback(net.IPv4(127, 0, 0, 1), net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	ll := NewLinkLayer(tap, loop)
	go ll.RxLoop(tap)
	go ll.RxLoop(loop)
	return ll
}
