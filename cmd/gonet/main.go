package main

import "github.com/mattcarp12/go-net/pkg/link"
import "github.com/mattcarp12/go-net/pkg/network"


var done = make(chan bool)

func main() {

	// Initialize the link layer
	link := link.Init()

	// Initialize the network layer
	network.Init(link)

	<-done
}
