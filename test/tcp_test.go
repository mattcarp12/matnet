package test

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"testing"

	s "github.com/mattcarp12/matnet/api"
	"github.com/mattcarp12/matnet/netstack"
	"github.com/stretchr/testify/assert"
)

const TCPPort = 8845

func LocalTCPAddr() string { return os.Getenv("LOCAL_IP") + ":" + fmt.Sprint(TCPPort) }

func startTCPServer(t *testing.T) *net.TCPListener {
	t.Helper()

	t.Logf("Starting TCP server on %s\n", LocalTCPAddr())

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(os.Getenv("LOCAL_IP")),
		Port: TCPPort,
		Zone: "",
	})
	assert.NoError(t, err)

	assert.Equal(t, LocalTCPAddr(), listener.Addr().String())

	return listener
}

func TestTCP_Connect(t *testing.T) {
	// Start TCP server
	listener := startTCPServer(t)
	defer listener.Close()

	// Create a new socket
	sock, err := s.Socket(s.SOCK_STREAM)
	assert.NoError(t, err)
	
	// Connect to the socket
	err = s.Connect(sock, LocalTCPAddr())
	assert.NoError(t, err)

	// Accept the connection on the server
	conn, err := listener.Accept()
	assert.NoError(t, err)

	// Get the remote addr
	remoteAddr := conn.RemoteAddr().String()
	t.Logf("Connected to %s\n", remoteAddr)
	assert.Equal(t, netstack.DefaultIPAddr, strings.Split(remoteAddr, ":")[0])
}

func TestTCP_SendClose(t *testing.T) {
	// Start TCP server
	listener := startTCPServer(t)
	defer listener.Close()

	// Create a new socket
	sock, err := s.Socket(s.SOCK_STREAM)
	assert.NoError(t, err)

	// Connect to the socket
	err = s.Connect(sock, LocalTCPAddr())
	assert.NoError(t, err)

	// Accept the connection on the server
	conn, err := listener.Accept()
	assert.NoError(t, err)

	// Initiate close
	err = s.Close(sock)
	assert.NoError(t, err)

	// Check if the connection is closed
	one := make([]byte, 1)
	_, err = conn.Read(one)
	assert.ErrorIs(t, err, io.EOF)
}