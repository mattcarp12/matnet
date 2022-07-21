package linklayer

import (
	"net"

	"github.com/mattcarp12/matnet/netstack"
	"github.com/mattcarp12/matnet/tuntap"
)

// Iface contains the common attributes for all network devices.
// Each device type should embed this struct.
type Iface struct {
	Name   string
	HwAddr net.HardwareAddr
	Mtu    uint16

	// Each interface maintains a list of addresses, since some may, for example,
	// have both an IPv4 and IPv6 address.
	IfAddrs []netstack.IfAddr

	// The type of L2 protocol that this interface supports.
	IfType netstack.ProtocolType

	tx_chan   chan *netstack.SkBuff
	LinkLayer *LinkLayer

	// TODO: Add interface statistics (low priority)
}

func (dev *Iface) GetType() netstack.ProtocolType {
	return dev.IfType
}

func (dev *Iface) TxChan() chan *netstack.SkBuff {
	return dev.tx_chan
}

func (dev *Iface) HandleRx(data []byte) {
	// Make a new SkBuff
	skb := netstack.NewSkBuff(data)

	// The skb is being passed to the link layer, so need to set the
	// type of packet so it can dispatch to the correct handler.
	skb.SetType(dev.GetType())

	// Set pointer to the NetworkInterface that this packet came in on
	skb.SetRxIface(dev)

	// Pass it to the link layer for further processing
	dev.LinkLayer.RxChan() <- skb
}

func (dev *Iface) GetHWAddr() net.HardwareAddr {
	return dev.HwAddr
}

func (iface *Iface) GetIfAddrs() []netstack.IfAddr {
	return iface.IfAddrs
}

func (iface *Iface) HasIPAddr(ip net.IP) bool {
	for _, ifAddr := range iface.IfAddrs {
		if ifAddr.IP.Equal(ip) {
			return true
		}
	}
	return false
}

func (dev *Iface) Read() ([]byte, error) {
	return nil, nil
}

func (dev *Iface) Write(data []byte) error {
	return nil
}

/******************************************************************************
	TAP Device
*******************************************************************************/

type TAPDevice struct {
	Iface
	tap *tuntap.Interface
}

func NewTap(tap *tuntap.Interface, name string, hwAddr net.HardwareAddr, addrs []netstack.IfAddr) *TAPDevice {
	netdev := TAPDevice{}
	netdev.Name = name
	netdev.tap = tap
	netdev.HwAddr = hwAddr
	netdev.IfAddrs = addrs
	netdev.IfType = netstack.ProtocolTypeEthernet
	netdev.tx_chan = make(chan *netstack.SkBuff)

	return &netdev
}

func (dev *TAPDevice) Read() ([]byte, error) {
	data := make([]byte, 1500)
	n, err := dev.tap.Read(data)
	return data[:n], err
}

func (dev *TAPDevice) Write(data []byte) error {
	_, err := dev.tap.Write(data)
	return err
}

/******************************************************************************
	Loopback device
*******************************************************************************/

type LoopbackDevice struct {
	Iface
	rx_chan chan []byte
}

func NewLoopback() *LoopbackDevice {
	netdev := LoopbackDevice{}
	netdev.Name = "lo"
	netdev.IfAddrs = []netstack.IfAddr{
		{
			IP:      net.IPv4(127, 0, 0, 1),
			Netmask: net.IPv4Mask(255, 255, 255, 255),
		},
	}
	netdev.HwAddr = net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	netdev.Mtu = 1500
	netdev.IfType = netstack.ProtocolTypeEthernet
	netdev.tx_chan = make(chan *netstack.SkBuff)
	netdev.rx_chan = make(chan []byte)
	return &netdev
}

// read from the write channel
func (dev *LoopbackDevice) Read() ([]byte, error) {
	skb := <-dev.TxChan()
	return skb.Data, nil
}

// and write to the read channel
func (dev *LoopbackDevice) Write(data []byte) error {
	dev.rx_chan <- data
	return nil
}
