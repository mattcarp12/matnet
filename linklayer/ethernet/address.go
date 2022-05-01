package ethernet

import "net"

func IsUnicast(addr net.HardwareAddr) bool {
	return addr[0]&0x01 == 0
}
