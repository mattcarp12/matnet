package netstack

import (
	"net"
)

type IfAddr struct {
	IP      net.IP
	Netmask net.IPMask
	Gateway net.IP
}

type NetworkInterface interface {
	Read() ([]byte, error)
	Write([]byte) error
	GetType() ProtocolType
	GetHWAddr() net.HardwareAddr
	GetIfAddrs() []IfAddr
	HasIPAddr(ip net.IP) bool

	// HandleRX is called when a packet is received from the "wire"
	HandleRx([]byte)

	// Network Devices have a TxChan that is used to send packets to the "wire"
	SkBuffWriter
}

func StartInterface(iface NetworkInterface) {
	go IfRxLoop(iface)
	go IfTxLoop(iface)
}

func IfRxLoop(iface NetworkInterface) {
	for {
		data, err := iface.Read()
		if err != nil {
			continue
		}
		iface.HandleRx(data)
	}
}

func IfTxLoop(iface NetworkInterface) {
	for {
		skb := <-iface.TxChan()
		if err := iface.Write(skb.Data); err != nil {
			skb.Error(err)
			continue
		}

		// Report that the packet was successfully sent on the wire
		skb.TxSuccess()
	}
}
