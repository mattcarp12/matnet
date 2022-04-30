package linklayer

import (
	"log"
	"net"

	"github.com/mattcarp12/go-net/netstack"
	"github.com/mattcarp12/go-net/tuntap"
)

// PhysicalDevice contains the common attributes for all network devices.
// Each device type should inherit this struct.
type PhysicalDevice struct {
	IP        net.IP
	MAC       net.HardwareAddr
	RxPackets uint64
	TxPackets uint64
	Type      netstack.ProtocolType
	txChan    chan *netstack.SkBuff
	RxHandler func([]byte)
	LinkLayer *LinkLayer
}

func (dev *PhysicalDevice) GetType() netstack.ProtocolType {
	return dev.Type
}

func (dev *PhysicalDevice) TxChan() chan *netstack.SkBuff {
	return dev.txChan
}

func (dev *PhysicalDevice) HandleRx(data []byte) {
	// Make a new SkBuff
	skb := netstack.NewSkBuff(data)

	// The skb is being passed to the link layer, so need to set the
	// type of packet so it can dispatch to the correct handler.
	skb.SetType(dev.GetType())

	// Set pointer to the NetworkInterface that this packet came in on
	skb.SetNetworkInterface(dev)

	// Pass it to the link layer for further processing
	dev.LinkLayer.RxChan() <- skb
}

func (dev *PhysicalDevice) GetHWAddr() net.HardwareAddr {
	return dev.MAC
}

func (dev *PhysicalDevice) GetNetworkAddr() net.IP {
	return dev.IP
}

func (dev *PhysicalDevice) Read() ([]byte, error) {
	return nil, nil
}

func (dev *PhysicalDevice) Write(data []byte) error {
	return nil
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
	netdev.Type = netstack.ProtocolTypeEthernet
	netdev.txChan = make(chan *netstack.SkBuff)

	return &netdev
}

func (dev *TAPDevice) Read() ([]byte, error) {
	data := make([]byte, 1500)
	n, err := dev.Iface.Read(data)
	return data[:n], err
}

func (dev *TAPDevice) Write(data []byte) error {
	log.Printf("TAPDevice.Write: %v", data)
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

func NewLoopback(ip net.IP, hw net.HardwareAddr) *LoopbackDevice {
	netdev := LoopbackDevice{}
	netdev.IP = ip
	netdev.MAC = hw
	netdev.Type = netstack.ProtocolTypeEthernet
	netdev.txChan = make(chan *netstack.SkBuff)
	netdev.rx_chan = make(chan []byte)
	return &netdev
}

func (dev *LoopbackDevice) Read() ([]byte, error) {
	skb := <-dev.TxChan()
	return skb.GetBytes(), nil
}

func (dev *LoopbackDevice) Write(data []byte) error {
	dev.rx_chan <- data
	return nil
}
