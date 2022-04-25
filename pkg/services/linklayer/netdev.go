package linklayer

import (
	"net"

	"github.com/mattcarp12/go-net/pkg/entities/netdev"
	"github.com/mattcarp12/go-net/pkg/entities/protocols"
	"github.com/mattcarp12/go-net/pkg/services/skbuff"
	"github.com/mattcarp12/go-net/pkg/tuntap"
)

type DataHandler func([]byte)

// PhysicalDevice contains the common attributes for all network devices.
// Each device type should inherit this struct.
type PhysicalDevice struct {
	IP        net.IP
	MAC       net.HardwareAddr
	RxPackets uint64
	TxPackets uint64
	Type      protocols.ProtocolType
	txChan    chan protocols.PDU
	RxHandler DataHandler
	LinkLayer *LinkLayer
}

func (dev *PhysicalDevice) GetType() protocols.ProtocolType {
	return dev.Type
}

func (dev *PhysicalDevice) TxChan() chan protocols.PDU {
	return dev.txChan
}

func (dev *PhysicalDevice) HandleRx(data []byte) {
	// Make a PDU
	skb := skbuff.New(data)

	// The PDU is being passed to the link layer, so need to set the
	// type of packet so it can dispatch to the correct handler.
	skb.SetType(dev.GetType())

	// Pass it to the link layer
	dev.LinkLayer.RxChan() <- skb
}

/**
TAP Device
*/

type TAPDevice struct {
	PhysicalDevice
	Iface *tuntap.Interface
}

func NewTap(iface *tuntap.Interface, ip net.IP, hw net.HardwareAddr) *TAPDevice {
	netdev := TAPDevice{}
	netdev.Iface = iface
	netdev.IP = ip
	netdev.MAC = hw
	return &netdev
}

func (dev *TAPDevice) Read() ([]byte, error) {
	data := make([]byte, 1500)
	n, err := dev.Iface.Read(data)
	return data[:n], err
}

func (dev *TAPDevice) Write(data []byte) error {
	_, err := dev.Iface.Write(data)
	return err
}

/**
Loopback device
*/

type LoopbackDevice struct {
	PhysicalDevice
	rx_chan chan []byte
}

var _ netdev.NetworkDevice = &LoopbackDevice{}

func NewLoopback(ip net.IP, hw net.HardwareAddr) *LoopbackDevice {
	netdev := LoopbackDevice{}
	netdev.IP = ip
	netdev.MAC = hw
	return &netdev
}

func (dev *LoopbackDevice) Read() ([]byte, error) {
	pdu := <-dev.TxChan()
	return pdu.GetBytes(), nil
}

func (dev *LoopbackDevice) Write(data []byte) error {
	dev.rx_chan <- data
	return nil
}
