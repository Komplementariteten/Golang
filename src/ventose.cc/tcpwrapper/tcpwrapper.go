package tcpwrapper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	INPUT_QUESIZE = 8
)

type TcpConfiguration struct {
	Port uint16
	TLS  bool
}

type TcpServer struct {
	listener *net.TCPListener
	clients  []*ClientConnection
}

type ClientConnection struct {
	client   *net.TCPConn
	ClientId int
}

func NewTcpWrapper(configuration *TcpConfiguration) (*TcpServer, error) {
	if configuration.TLS {
		panic("TLS is corrently not supported")
	}
	tcpresolver, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", configuration.Port))
	if err != nil {
		return nil, err
	}
	s := &TcpServer{}
	ln, err := net.ListenTCP(tcpresolver.Network(), tcpresolver)
	if err != nil {
		return nil, err
	}
	s.listener = ln
	return s, nil
}

func (s *TcpServer) Serve(response <-chan []byte) (request chan<- []byte, closeChan chan bool) {
	request = make(chan []byte, INPUT_QUESIZE)
	closeChan = make(chan bool)
	go func(localServer *TcpServer) {
		for {
			cc := &ClientConnection{}
			conn, err := localServer.listener.AcceptTCP()
			if err != nil {
				panic(err)
			}
			cc.client = conn
			go handleClientConnection(s, conn, request, response, closeChan)
		}
	}(s)
	return
}

func handleClientConnection(server *TcpServer, connection *net.TCPConn, request chan<- []byte, response <-chan []byte, quit <-chan bool) {

	var readBuffer bytes.Buffer
	for {
		select {
		case <-quit:
			binary.Write(connection, binary.LittleEndian, 23)
			close(request)
			return
		case responseData := <-response:
			connection.Write(responseData)
		}
		readBuffer.ReadFrom(connection)
		request <- readBuffer.Bytes()
		readBuffer.Reset()
	}
}
