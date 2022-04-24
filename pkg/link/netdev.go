package link

import (
	"github.com/mattcarp12/go-net/pkg/skbuff"
	"github.com/mattcarp12/go-net/pkg/tuntap"
	"net"
)

type NetworkDevice interface {
	Read(*skbuff.Sk_buff) error
	Write(*skbuff.Sk_buff) error
	Handle(*skbuff.Sk_buff) error
}

type FrameHandler func(*skbuff.Sk_buff) error

// PhysicalDevice contains the common attributes for all network devices.
// Each device type should inherit this struct.
type PhysicalDevice struct {
	IP        net.IP
	MAC       net.HardwareAddr
	RxPackets uint64
	TxPackets uint64
	Handler   FrameHandler
}

func (dev PhysicalDevice) Handle(skb *skbuff.Sk_buff) error {
	return dev.Handler(skb)
}

type TAPDevice struct {
	PhysicalDevice
	Iface *tuntap.Interface
}

func NewTap(iface *tuntap.Interface, ip net.IP, hw net.HardwareAddr) TAPDevice {
	netdev := TAPDevice{}
	netdev.Iface = iface
	netdev.IP = ip
	netdev.MAC = hw
	netdev.Handler = eth_rx_handle
	return netdev
}

func (dev TAPDevice) Read(skb *skbuff.Sk_buff) error {
	_, err := dev.Iface.Read(skb.Data)
	return err
}

func (dev TAPDevice) Write(skb *skbuff.Sk_buff) error {
	_, err := dev.Iface.Write(skb.Data)
	return err
}

type LoopbackDevice struct {
	PhysicalDevice
	rx_chan chan []byte
	tx_chan chan []byte
}

func NewLoopback(ip net.IP, hw net.HardwareAddr) LoopbackDevice {
	netdev := LoopbackDevice{}
	netdev.IP = ip
	netdev.MAC = hw
	netdev.Handler = eth_rx_handle
	return netdev
}

func (dev LoopbackDevice) Read(skb *skbuff.Sk_buff) error {
	skb.Data = <-dev.tx_chan
	return nil
}

func (dev LoopbackDevice) Write(skb *skbuff.Sk_buff) error {
	dev.rx_chan <- skb.Data
	return nil
}
