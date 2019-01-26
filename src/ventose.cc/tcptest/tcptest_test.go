package tcptest

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestLocalTcp(t *testing.T) {
	addr, err := net.ResolveTCPAddr("tcp6", "[::]:56784")
	if err != nil {
		t.Fatal(err)
	}
	ln, err := net.ListenTCP("tcp6", addr)
	defer ln.Close()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		con, err := ln.AcceptTCP()
		defer con.Close()
		if err != nil {
			t.Fatal(err)
		}
		con.Write([]byte("test9"))
		con.Write([]byte("test9"))
	}()

	client, err := net.DialTCP("tcp6", nil, addr)
	defer client.Close()

	if err != nil {
		t.Fatal(err)
	}
	data, err := bufio.NewReader(client).ReadBytes(byte('9'))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("Where is my Data")
	}
}

func TestLocalTcp_Scan(t *testing.T) {
	addr, err := net.ResolveTCPAddr("tcp6", "[::]:56784")
	if err != nil {
		t.Fatal(err)
	}
	ln, err := net.ListenTCP("tcp6", addr)
	defer ln.Close()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		con, err := ln.AcceptTCP()
		defer con.Close()
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 5; i++ {
			con.Write([]byte("test9"))
			con.Write([]byte{0, 0, 0})
		}
	}()

	client, err := net.DialTCP("tcp6", nil, addr)
	defer client.Close()

	if err != nil {
		t.Fatal(err)
	}
	delim := []byte{0, 0, 0}
	scanner := bufio.NewScanner(client)
	split := func(data []byte, atEOF bool) (adv int, token []byte, err error) {
		for i := len(delim); i < len(data); i++ {
			fragment := data[i-len(delim) : i]
			if bytes.Equal(fragment, delim) {
				return i, data[:i-len(delim)], nil
			}
		}
		size := len(data) - len(delim)
		return 0, data[:size], bufio.ErrFinalToken
	}
	scanner.Split(split)
	for scanner.Scan() {
		r := scanner.Bytes()
		if len(r) != 5 {
			t.Fatalf("Result has wrong size: %s", string(r))
		}
	}
}

func TestLocalTcp_AcceptLoop(t *testing.T) {
	max_runs := 10
	connected := 0
	addr, err := net.ResolveTCPAddr("tcp6", "[::]:56784")
	if err != nil {
		t.Fatal(err)
	}
	ln, err := net.ListenTCP("tcp6", addr)
	defer ln.Close()
	if err != nil {
		t.Fatal(err)
	}
	closeChan := make(chan bool)
	doneChan := make(chan bool)
	startedChan := make(chan bool)
	connectionCounter := make(chan bool, max_runs)
	delim := []byte{0, 0, 0}

	connectonHandle := func(con *net.TCPConn) {
		defer con.Close()
		con.SetDeadline(time.Now().Add(1 * time.Second))
		for i := 0; i < 2; i++ {
			data_ := append([]byte("test9"), delim...)
			_, err := con.Write(data_)
			if err != nil {
				t.Fatal(err)
				return
			}
		}
	}

	go func(ln *net.TCPListener) {
		for {
			con, err := ln.AcceptTCP()
			if err != nil {
				t.Fatal(err)
			}
			if con != nil {
				go connectonHandle(con)
			} else {
				return
			}
		}
	}(ln)


	go func(ln *net.TCPListener) {
		for {
			select {
			case <-closeChan:
				fmt.Printf("Closing: %d\n", connected)
				ln.Close()
				close(startedChan)
				close(connectionCounter)
				close(closeChan)
				doneChan <- true
				return
			case test := <-connectionCounter:
				if test {
					connected++
					startedChan <- true
				} else {
					connected--
				}
			}
		}
	}(ln)

	split := func(data []byte, atEOF bool) (adv int, token []byte, err error) {
		for i := len(delim); i < len(data); i++ {
			fragment := data[i-len(delim) : i]
			if bytes.Equal(fragment, delim) {
				return i, data[:i-len(delim)], nil
			}
		}
		if len(data) == 0 {
			return 0, data, bufio.ErrFinalToken
		}
		size := len(data) - len(delim)
		return 0, data[:size], bufio.ErrFinalToken
	}
	run_client := func(conChan chan bool) {
		client, err := net.DialTCP("tcp6", nil, addr)
		client.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Fatal(err)
		}
		scanner := bufio.NewScanner(client)
		scanner.Split(split)
		for scanner.Scan() {
			r := scanner.Bytes()
			if len(r) != 5 {
				t.Fatalf("Result has wrong size: %s", string(r))
			} else {
				t.Logf("%d - %s", connected, string(r))
			}
		}
		client.Close()
		conChan <- false
	}

	for i := 0; i < max_runs; i++ {
		connectionCounter <- true
		go run_client(connectionCounter)
	}
	go func(startedChan chan bool, closeChan chan bool) {
		<-startedChan
		for {
			if connected <= 0 {
				closeChan <- true
				return
			}
		}
	}(startedChan, closeChan)
	<-doneChan
}
