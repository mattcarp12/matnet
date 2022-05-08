package api

import "github.com/mattcarp12/go-net/netstack"

// Socket type constants
const (
	SOCK_STREAM = netstack.SocketTypeStream
	SOCK_DGRAM  = netstack.SocketTypeDatagram
	SOCK_RAW    = netstack.SocketTypeRaw
)

/*
	The user should only need to import the api package,
	so the following are re-exported.
*/

type SockAddr netstack.SockAddr
