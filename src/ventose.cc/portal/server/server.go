package server

import (
	"ventose.cc/pki"
	"sync"
	"ventose.cc/https"
	"ventose.cc/data"
	"ventose.cc/pki/certstore"
	"ventose.cc/portal/config"
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"ventose.cc/tools"
)

type PortalServer struct {
	storage       		*data.DataStorage
	certstore     		*certstore.CertStore
	pki 	      		*pki.PKI
	StorageWaiter 		sync.WaitGroup
	https	      		*https.HttpsFrontend
	S 	      		*config.PortalConfig
	StorageOnline 		bool
	PkiOnline     		bool
	StaticHttpOnline 	bool
	HttpsOnline 		bool
}

/*
 Internal Functions
 */

func (p *PortalServer) loadPki(cfg  *config.PkiConfiguration) (changed bool, err error) {
	if cfg == nil {
		return fmt.Errorf("PKI Configuration not found")
	}
	p.StorageWaiter.Wait()

	if !p.StorageOnline {
		return false, fmt.Errorf("No Storage availeable")
	}
	secure := tools.GetRandomBytes(32)
	if len(cfg.ID) == 0 {
		p.pki, err = pki.NewPki(p.S.Organisation, secure, cfg.RSAKeySize)
		if err != nil{
			return false, err
		}
		cs :=
	}
}

func (p *PortalServer) loadStorage(cfg  *data.InitialConfiguration) (err error) {
	p.StorageWaiter.Add(1)
	if cfg == nil {
		return fmt.Errorf("Database Configuration not found")
	}
	p.storage, err = data.OpenStorage(cfg)
	if err != nil {
		defer p.storage.Close()
		p.StorageOnline = true
		p.StorageWaiter.Done()
		p.storage.Serve()
	}
	return
}

func loadConfiguration(configFile string) (*config.PortalConfig, error) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, err
	}
	jb, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to Read Config file %s", configFile)
	}

	var cfg config.PortalConfig
	err = json.Unmarshal(jb, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to Unmarshall Configuration file %s", configFile)
	}
	return &cfg, nil
}

/*
 Package Initialisation
 */

func StartPortal(configFile string) (*PortalServer, error) {
	p := &PortalServer{
		StorageOnline: false,
		PkiOnline: false,
		HttpsOnline: false,
		StaticHttpOnline: false,
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("Portal Config File %s not found with: %v", configFile, err)
	}
	cfg, err := loadConfiguration(configFile)
	if err != nil {
		return nil, err
	}

	err = p.loadStorage(cfg.DB)

	if err != nil {
		return nil, fmt.Errorf("Failed to Load Storage with: %v", err)
	}

	return p, nil
}

