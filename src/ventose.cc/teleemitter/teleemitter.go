package teleemitter

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"log"
)

type MechanicStatus uint32

const (
	Moving MechanicStatus = iota
	Ready
	Blocked
	Setup
	NotConfigured
)

type OutgoingMessageInterface interface {
	Name() string
	Serialze() []byte
}

type IncommingMessageInterface interface {
	Serialize() []byte
	ParsePayload(b []byte) error
	Execute(m *Mechanic) (OutgoingMessageInterface, error)
}

type MessageParser interface {
	ValidateTypeCodeSize(typeCodeSize uint8) error
	Parse(b []byte) (IncommingMessageInterface, error)
}

type TeleEmitter struct {
	typeCodeSize uint8
	listener net.Listener
	err error
	processingQueue chan bytes.Buffer
	responseQueue chan OutgoingMessageInterface
	Active bool
	Mechanic *Mechanic
	messageParser MessageParser
	logger *log.Logger
}

type Mechanic struct {
	X int64
	Y int64
	Z int64
	Status MechanicStatus
}

func NewEmitter(tcpudp string, netString string, codeSize uint8, p MessageParser) (emitter *TeleEmitter, err error) {

	if strings.Compare(tcpudp, "tcp") != 0 || strings.Compare(tcpudp, "udp") != 0 {
		return nil, errors.New("can only be tcp or udp")
	}

	emitter = &TeleEmitter{}
	ln, err := net.Listen(tcpudp, netString)

	if err != nil {
		return nil, err
	}
	emitter.listener = ln
	emitter.typeCodeSize = codeSize
	emitter.messageParser = p
	emitter.logger = log.New(os.Stderr, "Log: ", log.Ldate | log.Ltime | log.Lshortfile)
	emitter.Mechanic = &Mechanic{
		X: 0,
		Y: 0,
		Z: 0,
		Status: NotConfigured,
	}
	return emitter, nil
}

func (n *TeleEmitter) Start() {
	go n.startProcessing()
	go func() {
		for {
			if !n.Active {
				return
			}
			conn, err := n.listener.Accept()
			if err != nil {
				n.err = err
				return
			}
			go handleConnection(conn, n)
		}
	} ()
}

func handleConnection(c net.Conn, emitter *TeleEmitter) {
	defer func() {
		if r := recover(); r != nil {
			emitter.Active = false
		}
	}()
	go handleRead(c, emitter)
	go handleWrite(c, emitter)
}

func handleRead(c net.Conn, emitter *TeleEmitter) {
	for {
		if !emitter.Active {
			return
		}
		data, err := ioutil.ReadAll(c)

		if err != nil {
			emitter.err = err
			panic(err)
		}
		buffer := bytes.Buffer{}
		buffer.Read(data)
		emitter.processingQueue <- buffer
	}
}

func handleWrite(c net.Conn, emitter *TeleEmitter) {
	for {
		if !emitter.Active {
			return
		}
		resp := <- emitter.responseQueue
		_, err := c.Write(resp.Serialze())
		if err != nil {
			panic(err)
		}
	}
}

func (n *TeleEmitter) startProcessing() {
	for val := range n.processingQueue {
		bytes := val.Bytes()
		msg, err := n.messageParser.Parse(bytes[0:n.typeCodeSize])
		if err != nil {
			n.logger.Println(err)
		}
		msg.ParsePayload(bytes[n.typeCodeSize:])
		resp, err := msg.Execute(n.Mechanic)
		if err != nil {
			n.logger.Println(err)
		} else {
			n.responseQueue <- resp
		}
	}
}

