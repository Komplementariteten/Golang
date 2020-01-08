package datastreamer

import (
	"bytes"
	"encoding/gob"
	"net"
	"testing"
	"time"
)

type teststruct struct {
	Id int64
	C1 []int64
	C2 []byte
}

func IgnoreCounter(tes *testing.T) chan int {
	counter := make(chan int)
	go func() {
		for {
			<-counter
		}
	}()
	return counter
}

func NewCounter(tes *testing.T) chan int {
	counter := make(chan int)
	go func(c chan int) {
		var total int64 = 0
		average := 0
		part := 0
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case n := <-c:
				total += int64(n)
				part += n
			case t := <-ticker.C:
				average = part / (t.Nanosecond() / 1000000)
				part = 0
				tes.Logf("%d Kb/s", average)
			}
		}
	}(counter)
	return counter
}

func NewTestClient(netstr string, waiter chan bool, size int64, t *testing.T) {
	var buff bytes.Buffer
	conn, err := net.Dial("tcp", netstr)
	defer conn.Close()

	b := make([]byte, 155)
	data := teststruct{}

	if err != nil {
		waiter <- false
		t.Fatal(err)
	}

	readen := int64(0)
	for {
		n, nerr := conn.Read(b)
		decoder := gob.NewDecoder(&buff)
		if nerr != nil {
			waiter <- false
			t.Fatal(nerr)
		}
		buff.Reset()
		buff.Write(b)

		derr := decoder.Decode(&data)
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

func TestNewStreamer(t *testing.T) {
	ctl := make(chan bool, 2)
	waiter := make(chan bool)
	out := IgnoreCounter(t)
	testdata := teststruct{}
	testdata.Id = 12
	testdata.C1 = make([]int64, 12)
	testdata.C2 = make([]byte, 64)

	netstr := "localhost:1224"
	NewGobStreamer(netstr, testdata, ctl, out)

	go NewTestClient(netstr, waiter, 155*1000, t)
	go NewTestClient(netstr, waiter, 155*1000, t)

	<-waiter
	<-waiter

	ctl <- true
}

func TestBenchmarkNewStreamer(t *testing.T) {
	ctl := make(chan bool, 2)
	waiter := make(chan bool)
	out := NewCounter(t)
	testdata := teststruct{}
	testdata.Id = 12
	testdata.C1 = make([]int64, 12)
	testdata.C2 = make([]byte, 64)

	netstr := "localhost:12233"
	NewGobStreamer(netstr, testdata, ctl, out)

	go NewTestClient(netstr, waiter, 155*10000000, t)
	go NewTestClient(netstr, waiter, 155*10000000, t)

	<-waiter
	<-waiter

	ctl <- true
}
