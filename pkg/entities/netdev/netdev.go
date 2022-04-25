package netdev

import (
	"log"
	"github.com/mattcarp12/go-net/pkg/entities/protocols"
)

type NetworkDevice interface {
	Read() ([]byte, error)
	Write([]byte) error
	GetType() protocols.ProtocolType

	// HandleRX is called when a packet is received from the "wire"
	HandleRx([]byte)

	// Network Devices have a TxChan that is used to send packets to the "wire"
	protocols.PDUWriter
}

func RxLoop(netdev NetworkDevice) {
	for {
		data, err := netdev.Read()
		if err != nil {
			log.Println("Error reading from network device:", err)
			continue
		}
		netdev.HandleRx(data)
	}
}

func TxLoop(netdev NetworkDevice) {
	for {
		pdu := <-netdev.TxChan()
		err := netdev.Write(pdu.GetBytes())
		if err != nil {
			log.Println("Error writing to network device:", err)
			continue
		}
	}
}
