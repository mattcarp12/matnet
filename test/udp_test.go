package test

import (
	"net"
	"os"
	"testing"

	s "github.com/mattcarp12/matnet/api"
	"github.com/mattcarp12/matnet/netstack"
	"github.com/stretchr/testify/assert"
)

const UDPPort = 8845

func startUDPServer(t *testing.T) *net.UDPConn {
	t.Helper()

	// create a new socket
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(os.Getenv("LOCAL_IP")),
		Port: UDPPort,
		Zone: "",
	})
	if err != nil {
		t.Fatalf("Error creating UDP socket: %v", err)
	}

	return conn
}

func readUDP(conn *net.UDPConn) string {
	buf := make([]byte, 1024)

	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		panic(err)
	}

	return string(buf[:n])
}

func makeUDPClient(t *testing.T) *net.UDPConn {
	t.Helper()

	// create a new client
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(netstack.DefaultIPAddr),
		Port: UDPPort,
		Zone: "",
	})
	assert.NoError(t, err)

	return conn
}

func TestUDPWrite(t *testing.T) {
	// Start Udp server
	conn := startUDPServer(t)
	defer conn.Close()

	// Create a new socket
	sock, err := s.Socket(s.SOCK_DGRAM)
	assert.NoError(t, err)

	// Make SockAddr to send to
	sockAddr := s.SockAddr{
		Port: UDPPort,
		IP:   net.ParseIP(os.Getenv("LOCAL_IP")),
	}

	// Send to the socket
	data := "Hello World\n"

	err = s.WriteTo(sock, []byte(data), 0, sockAddr)
	assert.NoError(t, err)

	// Check msg received by udp server
	if resp := readUDP(conn); resp != data {
		t.Errorf("Expected netcat output to be 'Hello World', got '%s'", resp)
	}
}

func TestUDPRead(t *testing.T) {
	client := makeUDPClient(t)
	defer client.Close()

	// Create a new socket
	sock, err := s.Socket(s.SOCK_DGRAM)
	assert.NoError(t, err)

	// Bind the socket
	err = s.Bind(sock, s.SockAddr{Port: UDPPort})
	assert.NoError(t, err)

	// Send data with client
	data := "Hello World\n"

	_, err = client.Write([]byte(data))
	assert.NoError(t, err)

	// Read data from socket
	buf := make([]byte, 1024)
	err = s.Read(sock, &buf)
	assert.NoError(t, err)

	t.Logf("Got data: %s", string(buf))

	// Check msg received by udp client
}
