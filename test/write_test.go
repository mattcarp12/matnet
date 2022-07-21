package test

import (
	"net"
	"os"
	"testing"
	"time"

	s "github.com/mattcarp12/matnet/api"
)

const UdpPort = 8845

func startUDPServer() *net.UDPConn {
	// create a new socket
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(os.Getenv("LOCAL_IP")),
		Port: UdpPort,
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

func TestWrite(t *testing.T) {
	// Start Udp server
	conn := startUDPServer()
	defer conn.Close()

	// Wait for setup
	time.Sleep(500 * time.Millisecond)

	// Create a new socket
	sock, err := s.Socket(s.SOCK_DGRAM)
	if err != nil {
		panic(err)
	}

	t.Logf("Created socket: %v", sock)

	// Make SockAddr to send to
	sockAddr := s.SockAddr{
		Port: UdpPort,
		IP:   net.ParseIP(os.Getenv("LOCAL_IP")),
	}

	// Send to the socket
	data := "Hello World\n"
	err = s.WriteTo(sock, []byte(data), 0, sockAddr)
	if err != nil {
		time.Sleep(500 * time.Millisecond)
		// Write again
		s.WriteTo(sock, []byte(data), 0, sockAddr)
	}

	// Check msg received by udp server
	if resp := readUDP(conn); resp != data {
		t.Errorf("Expected netcat output to be 'Hello World', got '%s'", resp)
	}

}
