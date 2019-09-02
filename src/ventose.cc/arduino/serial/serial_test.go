package serial

import (
	"testing"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

func TestSerialCommunication(t *testing.T) {
	options := serial.OpenOptions{
		PortName:        "/dev/cu.usbmodem1423201",
		BaudRate:        19200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		panic(err)
	}
	defer port.Close()
	for byteRate := 1; byteRate < 126; byteRate += 8 {
		b := make([]byte, byteRate)
		start := time.Now()
		for k := 0; k < 3000; k++ {
			_, err := port.Read(b)
			if err != nil {
				panic(err)
			}
			// go fmt.Printf("[%d]n = %v err = %v b = %v\n", k, n, err, b)
		}
		done := time.Now()
		delta := done.Sub(start)
		t.Logf("%f kb/s in %f s\n", float64(byteRate)/delta.Seconds(), delta.Seconds())
	}
}
