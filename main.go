package main

import (
	"github.com/mattcarp12/go-net/netstack/ipc"
	"github.com/mattcarp12/go-net/netstack/linklayer"
	"github.com/mattcarp12/go-net/netstack/networklayer"
	"github.com/mattcarp12/go-net/netstack/socket"
	"github.com/mattcarp12/go-net/netstack/transportlayer"
)

var done = make(chan bool)

func main() {

	// Initialize the link layer
	link, routing_table := linklayer.Init()

	// Initialize the network layer
	net := networklayer.Init(link)

	// Initialize the transport layer
	transport := transportlayer.Init(net)

	// Initialize the socket manager
	socket_layer := socket.Init(transport, routing_table)

	// Initialize the IPC server
	ipc.Init(socket_layer)

	<-done
}
