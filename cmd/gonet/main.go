package main

import (
	"github.com/mattcarp12/go-net/linklayer"
	"github.com/mattcarp12/go-net/networklayer"
)




var done = make(chan bool)

func main() {

	// Initialize the link layer
	link := linklayer.Init()

	// Initialize the network layer
	networklayer.Init(link)

	<-done
}
