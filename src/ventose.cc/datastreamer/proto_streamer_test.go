package datastreamer

import (
	"github.com/golang/protobuf/proto"
	"net"
	"testing"
)

func NewProtoTestClient(netstr string, waiter chan bool, size int64, t *testing.T) {
	conn, err := net.Dial("tcp", netstr)
	defer conn.Close()

	b := make([]byte, 457)
	data := &Content{}

	if err != nil {
		waiter <- false
		t.Fatal(err)
	}

	readen := int64(0)
	for {
		n, nerr := conn.Read(b)
		if nerr != nil {
			waiter <- false
			t.Fatal(nerr)
		}
		derr := proto.Unmarshal(b, data)
		if derr != nil {
			waiter <- false
			t.Fatal(derr)
		}

		readen += int64(n)
		if size == readen {
			waiter <- true
			t.Logf("%d (%d)\t", n, readen)
		}
	}
}

func GeneratreTestData() *Content {
	d := &Content{}
	var gates []*Gatesignal
	d.Fe = 1
	d.Ctp = 2345
	d.Channel = 1
	for i := 0; i < 50; i++ {
		g := &Gatesignal{}
		g.Amp = 1234
		g.GateATof = 110
		g.GateBTof = 120
		gates = append(gates, g)
	}
	d.Gate = gates
	return d
}

func TestNewProtoBStreamer(t *testing.T) {
	ctl := make(chan bool, 2)
	waiter := make(chan bool)
	out := IgnoreCounter(t)

	testdata := GeneratreTestData()

	netstr := "localhost:12255"
	NewProtoBStreamer(netstr, testdata, ctl, out)

	go NewProtoTestClient(netstr, waiter, 457*1000, t)
	go NewProtoTestClient(netstr, waiter, 457*1000, t)

	<-waiter
	<-waiter

	ctl <- true
}

func TestBenschmarkNewProtoBStreamer(t *testing.T) {
	ctl := make(chan bool, 2)
	waiter := make(chan bool)
	out := NewCounter(t)

	testdata := GeneratreTestData()

	netstr := "localhost:12256"
	NewProtoBStreamer(netstr, testdata, ctl, out)

	go NewProtoTestClient(netstr, waiter, 457*10000000, t)
	go NewProtoTestClient(netstr, waiter, 457*10000000, t)

	<-waiter
	<-waiter

	ctl <- true
}
