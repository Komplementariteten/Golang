package db

import (
	"log"
	"net"
	"net/rpc"
)

type DbRunner struct {
	db *Storage
}

func NewDb() {
	runner := new(DbRunner)
	s := rpc.NewServer()
	rpc.RegisterName(RPC_CONTROL_NAME, runner)
	l, e := net.Listen("tcp", RPC_TCP_PORT)
	if e != nil {
		log.Fatal("listener errort", e)
	}
	go s.Accept(l)
}

func (t *DbRunner) Start() {

}
