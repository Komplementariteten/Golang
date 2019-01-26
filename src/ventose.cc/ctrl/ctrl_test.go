package ctrl

import (
	"testing"
)

func TestNewConnection(t *testing.T) {
	cfg := &CtrlConfiguration{
		Port: 10109,
	}
	ctrl, err := NewConnection(cfg)
	if err != nil {

	}
}
