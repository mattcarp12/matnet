package netstack

import (
	"log"
	"net"
)

type NetworkInterface interface {
	Read() ([]byte, error)
	Write([]byte) error
	GetType() ProtocolType
	GetHWAddr() net.HardwareAddr
	GetNetworkAddr() net.IP
	GetNetmask() net.IPMask
	GetGateway() net.IP
	GetMTU() uint16

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
			log.Println("Error reading from network device:", err)
			continue
		}
		iface.HandleRx(data)
	}
}

func IfTxLoop(iface NetworkInterface) {
	for {
		skb := <-iface.TxChan()
		log.Printf("Sending packet to network device: %v", skb)
		if err := iface.Write(skb.GetBytes()); err != nil {
			log.Println("Error writing to network device:", err)
			skb.Error(err)
			continue
		}

		skb.TxSuccess(len(skb.GetBytes()))
		log.Printf("TX Success!")
	}
}
