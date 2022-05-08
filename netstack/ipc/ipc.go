package ipc

import (
	"bufio"
	"log"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/mattcarp12/go-net/netstack"
)

type IPC struct {
	socket_layer netstack.SocketLayer
	server       *IPCServer
	conn_map     map[string]*ipc_conn
	rx_chan      chan netstack.SockSyscallResponse
}

type ipc_conn struct {
	id           string
	conn         net.Conn
	socket_layer netstack.SocketLayer
	rx_chan      chan netstack.SockSyscallResponse
}

func (ic *ipc_conn) get_response() netstack.SockSyscallResponse {
	return <-ic.rx_chan
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

		// Generate unique connection ID
		conn_id := uuid.New().String()

		// Create new connection
		iconn := &ipc_conn{
			id:           conn_id,
			conn:         conn,
			socket_layer: ipc.socket_layer,
			rx_chan:      make(chan netstack.SockSyscallResponse),
		}

		// Add to connection map
		ipc.conn_map[iconn.id] = iconn

		// Start goroutine to handle connection
		go iconn.handle_connection()
	}
}

// SyscallResponseLoop ...
func (ipc *IPC) SyscallResponseLoop() {
	for {
		// get message from rx_chan
		msg := <-ipc.rx_chan
		log.Printf("Received message: %+v", msg)

		// get connection from map
		if conn, ok := ipc.conn_map[msg.ConnID]; ok {
			// send message to connection
			conn.rx_chan <- msg
		} else {
			log.Printf("Connection not found: %s", msg.ConnID)
		}
	}
}

func (server *IPCServer) Wait() {
	<-server.done
}

// handle_connection ...
// this is the goroutine that will handle the connection
func (iconn *ipc_conn) handle_connection() {
	reader := bufio.NewReader(iconn.conn)
	for {
		// Read request
		var req netstack.SockSyscallRequest
		err := req.Read(reader)
		if err != nil {
			log.Printf("Error reading request: %s", err)
			iconn.conn.Close()
			return
		}

		// Set the connection ID
		req.ConnID = iconn.id

		// Send request to socket layer
		iconn.socket_layer.SendSyscall(req)

		// Wait for response
		resp := iconn.get_response()
		log.Printf("Received response: %+v", resp)

		// Write response
		rawResp := append(resp.Bytes(), '\n')
		iconn.conn.Write(rawResp)
	}
}

func Init(sl netstack.SocketLayer) *IPC {
	os.Remove(ipc_addr)
	ipc := &IPC{
		sl,
		&IPCServer{
			make(chan bool),
		},
		make(map[string]*ipc_conn),
		make(chan netstack.SockSyscallResponse),
	}

	// Set SocketLayer rx_chan so it can sent messages to the IPC server
	sl.SetRxChan(ipc.rx_chan)

	go ipc.serve()
	go ipc.SyscallResponseLoop()
	return ipc
}
