package serverold

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"ventose.cc/auth"
	"ventose.cc/portal/server/graceful"
)

type HttpFrontendConfiguration struct {
	Port                 int
	Host                 string
	PublicStaticFilesDir string
	UploadDir            string
	SecStaticFilesDir    string
}

type HttpFrontend struct {
	Mux    *http.ServeMux
	Srv    *graceful.Server
	Portal *Portal
	Key    *rsa.PrivateKey
	Ln     net.Listener
	Logger *log.Logger
	auth.ConsumerInterface
	ab *auth.AuthBackend
}

//var SecStaticFilesConsumer *auth.Consumer = new(auth.Consumer)

func (hf *HttpFrontend) ConnectAuthBackend(ab *auth.AuthBackend) {
	hf.ab = ab
}

func (h *HttpFrontend) AddHandler(path string, hndlFunc http.HandlerFunc) error {
	h.Mux.HandleFunc(path, hndlFunc)
	return nil
}

func setUploadHandle(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func preHandleHandler(h http.HandlerFunc) http.HandlerFunc {
	return h
}

func setupDefaultHandlers(cfg *HttpFrontendConfiguration) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	if _, err := os.Stat(cfg.PublicStaticFilesDir); !os.IsPermission(err) || !os.IsNotExist(err) {
		log.Println("Adding Resorce Server")
		mux.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir(cfg.PublicStaticFilesDir))))
	} else {
		return nil, fmt.Errorf("Static files Dir %s is not accessable with %s", cfg.PublicStaticFilesDir, err.Error())
	}

	//mux.Handle("/auth/", http.StripPrefix("/auth/", ))

	/*if _, err := os.Stat(cfg.PublicStaticFilesDir); !os.IsPermission(err) || !os.IsNotExist(err) {
		mux.Handle("/files", preHandleHandler(setUploadHandle(cfg.UploadDir)))
	} else {
		return nil, fmt.Errorf("Static files Dir %s is not accessable with %s", cfg.PublicStaticFilesDir, err.Error())
	}*/
	if _, err := os.Stat(cfg.SecStaticFilesDir); !os.IsPermission(err) || !os.IsNotExist(err) {
		mux.Handle("/vault", preHandleHandler(setUploadHandle(cfg.SecStaticFilesDir)))
	}

	return mux, nil
}

func NewHttpFrontend(cfg *HttpFrontendConfiguration) (*HttpFrontend, error) {
	h := new(HttpFrontend)
	l, err := net.Listen("tcp6", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		log.Fatal(err)
	}
	h.Ln = l
	shutdown_callback := func() {
	}
	mux, err := setupDefaultHandlers(cfg)
	if err != nil {
		return nil, fmt.Errorf("[HttpServer] failed to setupDefault handler with:%s", err.Error())
	}
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: mux,
	}
	srv := &graceful.Server{
		Timeout:        1 * time.Second,
		BeforeShutdown: shutdown_callback,
		Server:         httpSrv,
	}
	h.Logger = new(log.Logger)
	h.Srv = srv
	return h, nil
}
