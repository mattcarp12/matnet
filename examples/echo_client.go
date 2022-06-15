package main

import (
	"fmt"
	"net"
	"time"

	s "github.com/mattcarp12/go-net/api"
)

func main() {
	// Create a new socket
	sock, err := s.Socket(s.SOCK_DGRAM)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created socket %+v\n", sock)

	// Make SockAddr to send to
	sock_addr := s.SockAddr{
		Port: 8845,
		IP:   net.IPv4(192, 168, 254, 17),
	}

	// Send to the socket
	err = s.WriteTo(sock, []byte("Hello World\n"), 0, sock_addr)
	if err != nil {
		time.Sleep(time.Second)
		// Write again
		s.WriteTo(sock, []byte("Hello World\n"), 0, sock_addr)
	}

	time.Sleep(time.Second * 5)

	// Read response from the socket
	buf := make([]byte, 1024)
	if err = s.Read(sock, &buf); err != nil {
		panic(err)
	}

	fmt.Printf("Read %d bytes: %s\n", len(buf), string(buf))

}
