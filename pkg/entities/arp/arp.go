package arp

import "net"

type ARPHeader struct {
	HardwareType uint16
	ProtocolType uint16
	HardwareSize uint8
	ProtocolSize uint8
	OpCode       uint16
	SenderMAC    net.HardwareAddr
	SenderIP     net.IP
	TargetMAC    net.HardwareAddr
	TargetIP     net.IP
}

type ARPCache map[string]net.HardwareAddr

const ARPTimeout = 5