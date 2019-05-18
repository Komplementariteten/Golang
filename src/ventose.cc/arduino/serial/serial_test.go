package serial

import (
	"fmt"
	"testing"
	"github.com/jacobsa/go-serial/serial"
)

func TestSerialCommunication(t *testing.T) {
	options := serial.OpenOptions{
		PortName: "/dev/cu.usbmodem1423201",
		BaudRate: 19200,
		DataBits: 8,
		StopBits: 1,
		MinimumReadSize: 4,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		panic(err)
	}
	defer port.Close()
	b := make([]byte, 1)
	for k := 0; k < 10000; k++ {
		n, err  := port.Read(b);
		fmt.Printf("[%d]n = %v err = %v b = %v\n", k, n, err, b)
	}
}
