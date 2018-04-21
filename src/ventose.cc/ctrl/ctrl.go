package ctrl

import (
	"net"
	"fmt"
	"time"
	"encoding/gob"
	"io"
)

type CtrlFunction func (m Message) string

type CtrlServer struct {
	Connection net.Listener
	Commands map[string]CtrlFunction
}

type CtrlConfiguration struct {
	Port int8
}

type Message struct {
	Command string
	Signatur string
	Text interface{}
	Time time.Time
}

func NewConnection(cf *CtrlConfiguration) (*CtrlServer, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", cf.Port))
	if err != nil {
		return nil, err
	}
	ctrl := &CtrlServer{
		Connection: ln,
	}
	go func(c *CtrlServer) {
		for {
			conn, err := c.Connection.Accept()
			if err != nil {
				c.error(err)
			} else {
				go c.handle(conn)
			}
		}
	}(ctrl)
	return ctrl, nil
}

func (c *CtrlServer) Close() {
	c.Connection.Close()
}

func (c *CtrlServer) handle(conn net.Conn) {
	var msg Message
	dec := gob.NewDecoder(conn)
	err := dec.Decode(&msg)
	if err != nil {
		c.error(err)
		return
	}
	if _, ok := c.Commands[msg.Command]; ok {
		resp := c.Commands[msg.Command](msg)
		io.WriteString(conn, resp)
	}
}

func (c *CtrlServer) error(e error){
	panic(e.Error())
}
