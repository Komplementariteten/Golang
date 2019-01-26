package ctrl

import (
	"encoding/json"
	"net/http"
	"time"
	"ventose.cc/data"
	"ventose.cc/tools"
	"ventose.cc/portal/server"
	"fmt"
	"ventose.cc/portal/config"
)

type AppHandle struct {
	Help           map[string]string
	PortalIntalled bool
}

type StatusResponse struct {
	ListenPort	uint
	CfgPath		string
	KeyPath		string
	CtrlRunning 	bool
	HttpRunning 	bool
	AuthRunning 	bool
	PortalRunning 	bool
}

func getStorageConfigFromParameter(w http.ResponseWriter, r *Request) (*data.InitialConfiguration, error) {
	storageCfg, found := r.Parameter["StorageSetup"]
	cfg := new(data.InitialConfiguration)
	if found {
		err := loadParameter(w, storageCfg, cfg)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}
	return nil, nil
}

func (h *AppHandle) LogCommand(r *Request) {
	logger.Printf("AppHandle - %s Parameters:%v", r.Command, r.Parameter)
}

func (h *AppHandle) StartPortalParts() {
	this.Portal.ServeStorage()
	this.Portal.ServeAuthBackend()
	this.Portal.ServeHttp(this.Cfg.StaticHttpConfig)
	this.Portal.ServePki(this.Cfg.PkiConfig)
	this.Portal.ServeHttps(this.Cfg.HttpsConfig)

}

func (h *AppHandle) HandleHttpState(r *Request, w http.ResponseWriter) {

}

func (h *AppHandle) HandlePortalState(r *Request, w http.ResponseWriter) {
	//logRequest(w,r)
	h.LogCommand(r)
	switch r.Command {
	case "shutdown":
		if this != nil && this.Portal != nil {
			this.Portal.Close()
		}
		if this != nil && this.Srv != nil {
			this.Srv.Stop(1 * time.Second)
		}
		w.Write(tools.ToBytes("OK"))
		SetupHeader(w)
	case "status":
		var state StatusResponse
		state.ListenPort  = this.Cfg.ListenPort
		state.KeyPath	  = this.Cfg.ServerKeyPath
		state.CfgPath	  = this.Cfg.CtrlServerConfig
		state.HttpRunning = this.Portal.StaticHttpOnline
		if this.Portal != nil && this.Portal.StorageOnline {
			state.PortalRunning = true
		} else {
			state.PortalRunning = false
		}

		jbytes, err := json.Marshal(state)
		if err != nil {
			http.Error(w, "Failed to Parse Configuration", http.StatusInternalServerError)
		}
		w.Write(jbytes)
		SetupHeader(w)
	case "start":
		defer func(){
			if r := recover(); r!= nil {
				fmt.Println("Recover ", r)
			}
		}()
		if this.Portal == nil {
			this.Portal = server_old.NewPortal()
			err := this.Portal.LoadStorage(this.Cfg.StorageConfiguration)
			if err != nil {
				http.Error(w, "Failed to start Storage on new Portal!", http.StatusInternalServerError)
			}
			h.StartPortalParts()
		} else if ! this.Portal.StorageOnline {
			err := this.Portal.LoadStorage(this.Cfg.StorageConfiguration)
			if err != nil {
				http.Error(w, "Failed to Load Storage!", http.StatusInternalServerError)
			}
			h.StartPortalParts()
		}
		w.Write(tools.ToBytes("OK"))
		SetupHeader(w)
	case "help":
		commandHelp := make(map[string]string)
		commandHelp["help"] = "Read this Message"
		commandHelp["shutdown"] = "Shut everything down, even the Control Server"
		commandHelp["setup"] = "Initial Setups Portal"

		jsonData, err := json.Marshal(commandHelp)
		if err != nil {
			http.Error(w, "Can't Parse Help List", http.StatusInternalServerError)
		}
		w.Write(jsonData)
		SetupHeader(w)
	default:
		http.Error(w, "Command Not Found", http.StatusNotFound)

	}

}
