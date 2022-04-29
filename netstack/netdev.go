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
	//TODO GetMTU() uint16

	// HandleRX is called when a packet is received from the "wire"
	HandleRx([]byte)

	// Network Devices have a TxChan that is used to send packets to the "wire"
	SkBuffWriter
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
		err := iface.Write(skb.GetBytes())
		if err != nil {
			log.Println("Error writing to network device:", err)
			continue
		}
	}
}
