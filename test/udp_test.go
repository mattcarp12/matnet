package test

import (
	"net"
	"os"
	"testing"

	s "github.com/mattcarp12/matnet/api"
)

const UDPPort = 8845

func startUDPServer() *net.UDPConn {
	// create a new socket
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(os.Getenv("LOCAL_IP")),
		Port: UDPPort,
		Zone: "",
	})
	if err != nil {
		panic(err)
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

func TestUDPWrite(t *testing.T) {
	// Start Udp server
	conn := startUDPServer()
	defer conn.Close()

	// Create a new socket
	sock, err := s.Socket(s.SOCK_DGRAM)
	if err != nil {
		panic(err)
	}

	t.Logf("Created socket: %v", sock)

	// Make SockAddr to send to
	sockAddr := s.SockAddr{
		Port: UDPPort,
		IP:   net.ParseIP(os.Getenv("LOCAL_IP")),
	}

	// Send to the socket
	data := "Hello World\n"

	err = s.WriteTo(sock, []byte(data), 0, sockAddr)
	if err != nil {
		t.Errorf("Error writing to socket: %v", err)
	}

	// Check msg received by udp server
	if resp := readUDP(conn); resp != data {
		t.Errorf("Expected netcat output to be 'Hello World', got '%s'", resp)
	}
}
