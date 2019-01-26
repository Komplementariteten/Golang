package tcpwrapper

import "testing"

func TerminationTest(t *testing.T) {
	cfg := &TcpConfiguration{}
	s, err := NewTcpWrapper(cfg)
	if err != nil {
		t.Fatal(err)
	}
	responseReader := make(chan []byte)
	_, closeChan := s.Serve(responseReader)
	closeChan <- true

}