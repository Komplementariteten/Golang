package auth

import (
	"crypto/rsa"
	"log"
	"net"
	"net/http"
	"ventose.cc/portal/server/graceful"
)

type AuthFrontend struct {
	Mux    *http.ServeMux
	Srv    *graceful.Server
	Key    *rsa.PrivateKey
	CtrlLn net.Listener
	Logger *log.Logger
}
