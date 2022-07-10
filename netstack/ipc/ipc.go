package ipc

import (
	"bufio"
	logging "log"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/mattcarp12/go-net/netstack"
	"github.com/mattcarp12/go-net/netstack/socket"
)

var log = logging.New(os.Stdout, "[IPC] ", logging.Ldate|logging.Lmicroseconds|logging.Lshortfile)

type IPC struct {
	socket_layer    *socket.SocketLayer
	server          *IPCServer
	conn_map        map[string]*ipc_conn
	SyscallRespChan chan netstack.SockSyscallResponse
}

type ipc_conn struct {
	id           string
	conn         net.Conn
	socket_layer *socket.SocketLayer
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
		// TODO : How does this get cleaned up?

		// Start goroutine to handle connection
		go iconn.handle_connection()
	}
}

// SyscallResponseLoop ...
// The IPC server received responses from the socket layer and dispatches
// them to the appropriate connection
func (ipc *IPC) SyscallResponseLoop() {
	for {
		// Get message from response channel and dispatch to connection
		// The response channel is actually the socket layer's syscall response channel
		msg := <-ipc.SyscallRespChan
		log.Printf("Received Syscall Response: %+v", msg)

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
// to the client process. It listens for Syscall Requests
// from the client and dispatches them to the socket layer.
func (iconn *ipc_conn) handle_connection() {
	reader := bufio.NewReader(iconn.conn)
	for {
		// Read request
		var req netstack.SockSyscallRequest
		err := req.Read(reader)
		if err != nil {
			log.Printf("Error reading request: %s", err)
			iconn.close()
			return
		}

		log.Printf("Received request: %+v", req)

		// Set the connection ID
		req.ConnID = iconn.id

		// Send request to socket layer
		iconn.socket_layer.SyscallReqChan <- req

		// Wait for response
		resp := iconn.get_response()

		// Write response
		rawResp := append(resp.Bytes(), '\n')
		iconn.conn.Write(rawResp)
	}
}

func (iconn *ipc_conn) close() {
	// make SockSyscallRequest for to close the socket
	req := netstack.SockSyscallRequest{
		ConnID:      iconn.id,
		SyscallType: netstack.SyscallClose,
	}

	// send request to socket layer
	iconn.socket_layer.SyscallReqChan <- req

	// wait for response
	resp := iconn.get_response()
	log.Printf("Received response: %+v", resp)

	// close connection
	iconn.conn.Close()
}

func Init(sl *socket.SocketLayer) *IPC {
	os.Remove(ipc_addr)
	ipc := &IPC{
		sl,
		&IPCServer{
			make(chan bool),
		},
		make(map[string]*ipc_conn),
		make(chan netstack.SockSyscallResponse),
	}

	// Set SocketLayer syscall response channel so it can
	// send messages to the IPC server
	sl.SyscallRespChan = ipc.SyscallRespChan

	go ipc.serve()
	go ipc.SyscallResponseLoop()
	return ipc
}
