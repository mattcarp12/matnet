package ipc

import (
	"bufio"
	"log"
	"net"
	"os"

	"github.com/mattcarp12/go-net/netstack"
)

type IPC struct {
	socket_manager netstack.SocketManager
	server         *IPCServer
}

// IPCServer ...
type IPCServer struct {
	done chan bool
}

const ipc_addr = "/tmp/gonet.sock"

// serve ...
func (ipc *IPC) serve() {
	log.Printf("Starting server on %s", ipc_addr)

	listener, err := net.Listen("unix", ipc_addr)
	if err != nil {
		log.Fatal(err)
	}

	// change file permission so non-root users can access
	os.Chmod(ipc_addr, 0777)

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		}

		go ipc.handle_connection(conn)
	}
}

func (server *IPCServer) Wait() {
	<-server.done
}

func (ipc *IPC) handle_connection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		// Read request
		bytes, err := reader.ReadBytes('\n')
		if err != nil {
			log.Printf("Error reading bytes: %s", err)
			break
			// TODO: Handle error somehow
		}

		// Parse message
		msg, err := ParseMsg(bytes)
		if err != nil {
			log.Printf("Error parsing message: %s", err)
			continue
		}

		var res interface{}

		// Handle command
		switch msg.Command {
		case SOCKET:
			res = ipc.socket(msg.Data)
		case CONNECT:
			res = ipc.connect(msg.Data)
		case WRITETO:
			res = ipc.writeto(msg.Data)
		default:
			log.Printf("Unknown command: %s", msg.Command)

		}

		// Marshal the response into a byte array
		data := MakeMsg(msg.Command, res)

		// Send response
		log.Printf("Sending response: %+v", res)
		conn.Write(data)
	}
}

func Init(sm netstack.SocketManager) *IPC {
	os.Remove(ipc_addr)
	ipc := &IPC{
		sm,
		&IPCServer{
			make(chan bool),
		}}
	go ipc.serve()
	return ipc
}
