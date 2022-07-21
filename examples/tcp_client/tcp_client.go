package main

import (
	"flag"
	"net"

	s "github.com/mattcarp12/matnet/api"
)

func main() {
	// Make destination address
	destAddr := flag.String("dest", "", "Destination address")
	flag.Parse()

	// create new socket
	sock, err := s.Socket(s.SOCK_STREAM)
	if err != nil {
		panic(err)
	}

	ip := net.ParseIP(*destAddr)
	if ip == nil {
		panic("Invalid IP address")
	}

	dest := s.SockAddr{
		Port: 8845,
		IP:   ip,
	}

	// Connect to the socket
	if err = s.Connect(sock, dest); err != nil {
		panic(err)
	}

	// Write to the socket
	if err = s.Write(sock, []byte("Hello World\n"), 0); err != nil {
		panic(err)
	}
}
