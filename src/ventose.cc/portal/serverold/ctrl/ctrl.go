package ctrl

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"time"
	"ventose.cc/data"
	"ventose.cc/pki"
	"ventose.cc/portal/config"
	"ventose.cc/portal/server"
	"ventose.cc/portal/server/graceful"
	"ventose.cc/portal/serverold"
	"ventose.cc/tools"
)

const (
	MIMEREGEX = ".*/json.*"
)

type Request struct {
	Command   string
	Signature []byte
	Parameter map[string]string
	Date      time.Time
}

type ControllServerConfiguration struct {
	ListenPort           uint
	ServerKeyPath        string
	StorageConfiguration *data.InitialConfiguration
	CtrlServerLogFile    string
	CtrlServerConfig     string
	StaticHttpConfig     *serverold.HttpFrontendConfiguration
	HttpsConfig          *config.PortalHttpsConfiguration
	PkiConfig            *pki.PKIConfiguration
}

type ControllServer struct {
	Mux    *http.ServeMux
	Srv    *graceful.Server
	Cfg    *ControllServerConfiguration
	Portal *server_old.Portal
	Key    *rsa.PrivateKey
	CtrlLn net.Listener
	Logger *log.Logger
}

type HandlerFunc func(*Request, http.ResponseWriter)

var this *ControllServer
var logger *log.Logger
var logFile *os.File
var mimeRegex *regexp.Regexp

func gracefulShutdown() {
	logger.Println("Shutdown")
	logFile.Close()
}

func (c *ControllServer) SetupHandlers() *http.ServeMux {

	mux := http.NewServeMux()

	defaulthndl := new(CtrlHandle)

	setup := new(AppHandle)

	mux.HandleFunc("/portal/state", validateRequest(setup.HandlePortalState))
	mux.HandleFunc("/portal/http", validateRequest(setup.HandleHttpState))
	mux.HandleFunc("/", validateRequest(defaulthndl.Default))
	return mux
}

func GetControllServerDefaultConfig() (*ControllServerConfiguration, error) {
	ctrlCfg := new(ControllServerConfiguration)
	ctrlCfg.CtrlServerLogFile = "/var/log/portal.log"
	ctrlCfg.ListenPort = 8918
	ctrlCfg.CtrlServerConfig = "/etc/portal/portal.json"
	ctrlCfg.ServerKeyPath = "/etc/portal/"
	ctrlCfg.StorageConfiguration = new(data.InitialConfiguration)
	ctrlCfg.StorageConfiguration.ConfigPath = "/etc/portal/db.cfg"
	ctrlCfg.StorageConfiguration.DbPath = "/etc/portal/db/"
	ctrlCfg.StorageConfiguration.MaxDatabases = 12
	ctrlCfg.StorageConfiguration.PathAccessMode = 0700
	ctrlCfg.StorageConfiguration.AuthKey = tools.GetRandomAsciiString(64)
	ctrlCfg.StaticHttpConfig = new(serverold.HttpFrontendConfiguration)
	ctrlCfg.StaticHttpConfig.Port = 8080
	ctrlCfg.StaticHttpConfig.Host = "localhost"
	return ctrlCfg, nil
}

func (cfg *ControllServerConfiguration) SetBasePath(path string) error {

	var err error
	var fi os.FileInfo
	if fi, err = os.Stat(path); os.IsPermission(err) || !fi.IsDir() || os.IsNotExist(err) {
		return fmt.Errorf("Can't access %s, you do not have access or it is no Directory", err)
	}

	logDir := path + tools.OSSP + "log"

	err = os.Mkdir(logDir, 0700)
	if err != nil {
		return fmt.Errorf("Failed to create Log Directory: %s", err)
	}
	cfg.CtrlServerLogFile = logDir + tools.OSSP + "portal.log"

	configDir := path + tools.OSSP + "config"
	err = os.Mkdir(configDir, 0700)
	if err != nil {
		return fmt.Errorf("Failed to create Config Directory: %s", err)
	}

	cfg.StorageConfiguration.ConfigPath = configDir + tools.OSSP + "db.cfg"
	cfg.ServerKeyPath = configDir
	cfg.CtrlServerConfig = configDir + tools.OSSP + "portal.json"
	dbDir := path + tools.OSSP + "db"
	err = os.Mkdir(dbDir, 0700)
	if err != nil {
		return fmt.Errorf("Failed to create Database Directory: %s", err)
	}
	cfg.StorageConfiguration.DbPath = dbDir

	httpDir := path + tools.OSSP + "htdocs"
	err = os.Mkdir(httpDir, 0777)
	if err != nil {
		return fmt.Errorf("Failed to create Http Directory: %s", err)
	}
	cfg.StaticHttpConfig.PublicStaticFilesDir = httpDir

	httpsDir := path + tools.OSSP + "htsdocs"
	err = os.Mkdir(httpsDir, 0700)
	if err != nil {
		return fmt.Errorf("Failed to create Https Directory: %s", err)
	}
	cfg.StaticHttpConfig.SecStaticFilesDir = httpsDir

	filesDir := path + tools.OSSP + "upload"
	err = os.Mkdir(filesDir, 0700)
	if err != nil {
		return fmt.Errorf("Failed to create Http Upload Directory: %s", err)
	}
	cfg.StaticHttpConfig.UploadDir = filesDir

	return nil

}

func GetControlServer(cfg *ControllServerConfiguration) *ControllServer {

	mimeRegex = regexp.MustCompile(MIMEREGEX)

	ctrlSrv := new(ControllServer)
	ctrlSrv.Cfg = cfg

	//Logging
	var err error
	logFile, err = os.OpenFile(cfg.CtrlServerLogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	ctrlSrv.Logger = log.New(logFile, "[ControlServer] ", log.Ldate|log.Ltime)

	ctrlSrv.Logger.Println("Starting ControllServer")
	log.SetOutput(logFile)
	// RSA KEy
	rsaFile := new(os.File)
	key := new(rsa.PrivateKey)
	if fi, err := os.Stat(cfg.ServerKeyPath); os.IsPermission(err) {
		ctrlSrv.Logger.Panicf("ControlServer is not allowed to Access PrivateKeyFile %s\n", cfg.ServerKeyPath)
	} else if os.IsNotExist(err) {
		// Open File
		rsaFile, err = os.OpenFile(cfg.ServerKeyPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
		if err != nil {
			ctrlSrv.Logger.Panic(err)
		}
		defer rsaFile.Close()

		key, err = rsa.GenerateKey(rand.Reader, 2056)
		if err != nil {
			ctrlSrv.Logger.Panic(err)
		}
		bytes := x509.MarshalPKCS1PrivateKey(key)
		pemBlock := new(pem.Block)
		pemBlock.Bytes = bytes
		pemBlock.Type = "RSA PRIVATE KEY"

		err = pem.Encode(rsaFile, pemBlock)

		if err != nil {
			ctrlSrv.Logger.Printf("Can't save Controll Server Private Key to %s with: %v\n", cfg.ServerKeyPath, err)
		}
		ctrlSrv.Logger.Println("RSA Key generated")

	} else if err == nil {

		mode := fi.Mode()
		if _, err := os.Open(fmt.Sprintf("%s%srsa.key", cfg.ServerKeyPath, tools.GetDirSeperator())); mode.IsDir() && os.IsNotExist(err) {
			// Test Directory Seperator
			rsaFile, err = os.OpenFile(fmt.Sprintf("%s%srsa.key", cfg.ServerKeyPath, tools.GetDirSeperator()), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
			if err != nil {
				ctrlSrv.Logger.Panic(err)
			}
			defer rsaFile.Close()
			key, err = rsa.GenerateKey(rand.Reader, 1024)
			if err != nil {
				ctrlSrv.Logger.Panic(err)
			}
			bytes := x509.MarshalPKCS1PrivateKey(key)
			pemBlock := new(pem.Block)
			pemBlock.Bytes = bytes
			pemBlock.Type = "RSA PRIVATE KEY"

			err = pem.Encode(rsaFile, pemBlock)

			if err != nil {
				ctrlSrv.Logger.Panicf("Can't save Control Server Private Key with: %v\n", err)
			}
			ctrlSrv.Logger.Println("RSA Key generated")

		} else {
			err = nil
			if mode.IsDir() {
				rsaFile, err = os.Open(fmt.Sprintf("%s%srsa.key", cfg.ServerKeyPath, tools.GetDirSeperator()))
			} else {
				rsaFile, err = os.Open(cfg.ServerKeyPath)
			}
			if err != nil {
				log.Panic(err)
			}
			defer rsaFile.Close()
			buffer := new(bytes.Buffer)
			n, err := buffer.ReadFrom(rsaFile)

			if err != nil {
				ctrlSrv.Logger.Panic(err)
			}

			if n == 0 {
				ctrlSrv.Logger.Panicln("No Data read from RSA Key file")
			}

			keyBytes, _ := pem.Decode(buffer.Bytes())

			key, err = x509.ParsePKCS1PrivateKey(keyBytes.Bytes)
			if err != nil {
				ctrlSrv.Logger.Panicf("Can't read RSA Key from %s with: %v\n", cfg.ServerKeyPath, err)
			}
			ctrlSrv.Logger.Println("RSA Key loaded")
		}

	} else {
		ctrlSrv.Logger.Panic(err)
	}
	ctrlSrv.Key = key

	this = ctrlSrv
	logger = ctrlSrv.Logger

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", cfg.ListenPort))
	if err != nil {
		log.Fatal(err)
	}
	ctrlSrv.CtrlLn = l

	shutdown_callback := func() {
		gracefulShutdown()
	}

	mux := ctrlSrv.SetupHandlers()
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", cfg.ListenPort),
		Handler: mux,
	}
	srv := &graceful.Server{
		Timeout:        1 * time.Second,
		BeforeShutdown: shutdown_callback,
		Server:         httpSrv,
	}

	ctrlSrv.Srv = srv
	return ctrlSrv
}

func StartControllServer(cfg *ControllServerConfiguration) chan bool {
	srv := GetControlServer(cfg)
	waiter := make(chan bool)
	go func() {
		err := srv.Srv.Serve(srv.CtrlLn)
		if err != nil {
			panic(err)
		}
		waiter <- true
	}()
	return waiter
}
