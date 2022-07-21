package api

import (
	"github.com/mattcarp12/matnet/netstack"
	"github.com/mattcarp12/matnet/netstack/socket"
)

// Socket type constants
const (
	SOCK_STREAM = socket.SocketTypeStream
	SOCK_DGRAM  = socket.SocketTypeDatagram
	SOCK_RAW    = socket.SocketTypeRaw
)

/*
	The user should only need to import the api package,
	so the following are re-exported.
*/

type SockAddr netstack.SockAddr
