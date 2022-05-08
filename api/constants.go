package api

import "github.com/mattcarp12/go-net/netstack"

// Socket domain constants
const (
	AF_UNIX  = 1
	AF_INET  = 2
	AF_INET6 = 3
)

// Socket type constants
const (
	SOCK_STREAM    = netstack.SocketTypeStream
	SOCK_DGRAM     = netstack.SocketTypeDatagram
	SOCK_RAW       = netstack.SocketTypeRaw
)

// Socket protocol constants

/*
	The user should only need to import the api package,
	so the following are re-exported.
*/

type SockAddr netstack.SockAddr
