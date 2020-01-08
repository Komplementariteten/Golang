package datastreamer

import (
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"os"
	"runtime"
)

func NewProtoBStreamer(netstring string, data proto.Message, ctrl <-chan bool, counter chan int) chan bool {
	var stop = false
	graceful := make(chan bool)
	netbytes, err := proto.Marshal(data)
	if err != nil {
		log.Fatal("encode:", err)
	}

	ln, err := net.Listen("tcp", netstring)

	if err != nil {
		log.Fatal("listen:", err)
	}

	go func() {
		shutdown := <-ctrl
		stop = true
		ln.Close()
		if !shutdown {
			graceful <- false
			os.Exit(5)
		}
	}()

	go func(l net.Listener, buff []byte, count chan int) {
		defer l.Close()
		for {
			conn, _ := l.Accept()
			defer conn.Close()
			go func(c net.Conn, b []byte, out chan int) {
				for {
					n, _ := c.Write(b)
					out <- n
					if stop {
						graceful <- false
						runtime.Goexit()
					}
				}
				graceful <- true
			}(conn, buff, count)
		}
	}(ln, netbytes, counter)
	return graceful
}
