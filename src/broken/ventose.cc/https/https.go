package https

import (
	"crypto/tls"
	"net/http"
	"fmt"
	"net"
	"os"
	"strings"
)

type HttpsConfiguration struct {
	Address 	string
	StaticPath	string
}

type HttpsFrontend struct {
	server 		*http.Server
	mux 		*http.ServeMux
	ln		net.Listener
	IsListening 	bool
	IsReady		bool
	cfg		*HttpsConfiguration
}

func defaultHttpsFrontend() *HttpsFrontend {
	frontend := new(HttpsFrontend)
	frontend.IsListening = false
	frontend.IsReady = false
	return frontend
}

func NewHttpsFrontend(cert tls.Certificate, cfg *HttpsConfiguration) (*HttpsFrontend, error){
	tlsconfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},
	}

	mux := http.NewServeMux()
	f := defaultHttpsFrontend()
	var e error
	f.ln, e = tls.Listen("tcp6", cfg.Address, tlsconfig)
	if e != nil {
		return nil, e
	}
	f.mux = mux
	f.server = &http.Server{
		Addr: cfg.Address,
		TLSConfig: tlsconfig,
		Handler: mux,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	f.cfg = cfg
	return f, nil
}

func (f *HttpsFrontend) Handle(path string, fn http.HandlerFunc) error {
	f.mux.HandleFunc(path, fn)
	f.IsReady = true
	return nil
}

func (f *HttpsFrontend) Run() error {
	if !f.IsReady {
		return fmt.Errorf("No Path Handlers are Set, exiting")
	}
	if _, err := os.Stat(f.cfg.StaticPath); !os.IsNotExist(err) {
		f.Handle("/resource/", func(w http.ResponseWriter, r *http.Request){
			fPath := strings.Replace(r.RequestURI, "/resource", f.cfg.StaticPath, 1)
			http.ServeFile(w, r, fPath)
		})
	}
	var err error
	go func(err error){
		err = f.server.Serve(f.ln)
	}(err)
	if err != nil {
		f.IsListening = true
		return nil
	}

	return err
}
