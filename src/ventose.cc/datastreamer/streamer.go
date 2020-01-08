package datastreamer

import (
	"bytes"
	"encoding/gob"
	"log"
	"net"
	"os"
	"runtime"
)

func NewGobStreamer(netstring string, data interface{}, ctrl <-chan bool, counter chan int) {
	var netbytes bytes.Buffer
	enc := gob.NewEncoder(&netbytes)
	err := enc.Encode(data)
	var stop = false
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
			os.Exit(5)
		}
	}()

	go func(l net.Listener, buff bytes.Buffer, count chan int) {
		defer l.Close()
		for {
			conn, _ := l.Accept()
			/* if err != nil {
				log.Fatal(err)
			} */
			defer conn.Close()
			go func(c net.Conn, b bytes.Buffer, out chan int) {
				outbytes := b.Bytes()
				for {
					n, _ := c.Write(outbytes)
					out <- n
					if stop {
						runtime.Goexit()
					}
				}
			}(conn, buff, count)
		}
	}(ln, netbytes, counter)
}
