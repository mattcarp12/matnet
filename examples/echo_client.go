package main

import (
	"fmt"
	"net"

	s "github.com/mattcarp12/go-net/api"
)

func main() {
	// Create a new socket
	sock, err := s.Socket(s.AF_INET, s.SOCK_DGRAM, 0)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created socket %+v\n", sock)

	// Make SockAddr to send to
	sock_addr := s.SockAddr{
		Port: 45880,
		IP:   net.IPv4(192, 168, 254, 17),
	}

	// Send to the socket
	err = s.WriteTo(sock, []byte("Hello World"), 0, sock_addr)

}
