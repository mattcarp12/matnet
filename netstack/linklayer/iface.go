package linklayer

import (
	"net"

	"github.com/mattcarp12/go-net/netstack"
	"github.com/mattcarp12/go-net/tuntap"
)

type IfaceConfig struct {
	Name    string
	IP      net.IP
	MAC     net.HardwareAddr
	Netmask net.IPMask
	Gateway net.IP
	Mtu     uint16
}

// Iface contains the common attributes for all network devices.
// Each device type should inherit this struct.
type Iface struct {
	IfaceConfig
	RxPackets uint64
	TxPackets uint64
	Type      netstack.ProtocolType
	txChan    chan *netstack.SkBuff
	RxHandler func([]byte)
	LinkLayer *LinkLayer
}

func (dev *Iface) GetType() netstack.ProtocolType {
	return dev.Type
}

func (dev *Iface) TxChan() chan *netstack.SkBuff {
	return dev.txChan
}

func (dev *Iface) HandleRx(data []byte) {
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

func (dev *Iface) GetHWAddr() net.HardwareAddr {
	return dev.MAC
}

func (dev *Iface) GetNetworkAddr() net.IP {
	return dev.IP
}

func (dev *Iface) GetNetmask() net.IPMask {
	return dev.Netmask
}

func (dev *Iface) GetGateway() net.IP {
	return dev.Gateway
}

func (dev *Iface) GetMTU() uint16 {
	return dev.Mtu
}

func (dev *Iface) Read() ([]byte, error) {
	return nil, nil
}

func (dev *Iface) Write(data []byte) error {
	return nil
}

/**
TAP Device
*/

type TAPDevice struct {
	Iface
	tap *tuntap.Interface
}

func NewTap(tap *tuntap.Interface, config IfaceConfig) *TAPDevice {
	netdev := TAPDevice{}
	netdev.tap = tap
	netdev.IfaceConfig = config
	netdev.Type = netstack.ProtocolTypeEthernet
	netdev.txChan = make(chan *netstack.SkBuff)

	return &netdev
}

func (dev *TAPDevice) Read() ([]byte, error) {
	data := make([]byte, 1500)
	n, err := dev.tap.Read(data)
	return data[:n], err
}

func (dev *TAPDevice) Write(data []byte) error {
	log.Printf("TAPDevice.Write: %v", data)
	_, err := dev.tap.Write(data)
	return err
}

/**
Loopback device
*/

type LoopbackDevice struct {
	Iface
	rx_chan chan []byte
}

func NewLoopback() *LoopbackDevice {
	netdev := LoopbackDevice{}
	netdev.Name = "lo"
	netdev.IP = net.IPv4(127, 0, 0, 1)
	netdev.MAC = net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	netdev.Netmask = net.IPv4Mask(255, 0, 0, 0)
	netdev.Mtu = 1500
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
