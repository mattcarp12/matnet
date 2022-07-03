package test

import (
	"net"
	"os/exec"
	"testing"
	"time"

	s "github.com/mattcarp12/go-net/api"
	"github.com/mattcarp12/go-net/netstack/ipc"
	"github.com/mattcarp12/go-net/netstack/linklayer"
	"github.com/mattcarp12/go-net/netstack/networklayer"
	"github.com/mattcarp12/go-net/netstack/socket"
	"github.com/mattcarp12/go-net/netstack/transportlayer"
)

var done = make(chan bool)

func StartNetstack() {
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

func StopNetstack() {
	done <- true
}

var nc *exec.Cmd

func StartNetcatServer() {
	nc = exec.Command("nc", "-l", "-u", "8845")
	nc.Start()
}

func StopNetcatServer() {
	nc.Process.Kill()
}

func GetNetcatOutput() string {
	buf := make([]byte, 1024)
	stdout, _ := nc.StdoutPipe()
	stdout.Read(buf)
	return string(buf)
}

func TestWrite(t *testing.T) {
	// Start the network stack
	go StartNetstack()
	defer StopNetstack()

	// Start netcat server
	go StartNetcatServer()
	defer StopNetcatServer()

	// Wait for setup
	time.Sleep(time.Second)

	// Create a new socket
	sock, err := s.Socket(s.SOCK_DGRAM)
	if err != nil {
		panic(err)
	}

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

	// Check netcat output
	res := GetNetcatOutput()
	t.Logf("Netcat output: %s", res)
	if res != "Hello World\n" {
		t.Errorf("Expected netcat output to be 'Hello World\n', got '%s'", res)
	}
}
