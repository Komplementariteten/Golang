package serverold

import (
	"fmt"
	"sync"
	"ventose.cc/auth"
	"ventose.cc/data"
	"ventose.cc/pki/certstore"
	"ventose.cc/pki"
	"ventose.cc/https"
	"math/big"
	"ventose.cc/portal/config"
)


type Portal struct {
	storage       *data.DataStorage
	certstore     *certstore.CertStore
	pki 	      *pki.PKI
	StorageWaiter sync.WaitGroup
	authbackend   *auth.AuthBackend
	http	      *HttpFrontend
	https	      *https.HttpsFrontend
	S 	      *config.PortalConfig
	StorageOnline bool
	PkiOnline     bool
	StaticHttpOnline bool
	HttpsOnline bool
}

func NewPortal() *Portal {
	p := new(Portal)
	p.StorageWaiter.Add(1)
	p.StorageOnline = false
	return p
}

func (p *Portal) ServePki(cfg *pki.PKIConfiguration) error {
	if !p.StorageOnline {
		return fmt.Errorf("First you must start the Database Storage")
	}

	if cfg.PKIId != nil {
		err := p.certstore.LoadCertStore(cfg.PKIId)
		return err
	} else {
		err := p.certstore.New("Portal CertStore", cfg.Organization, pki.DEFAULT_ASYMKEYLEN)
		if err != nil {
			return err
		}
		cfg.PKIId = p.certstore.Config.Id
	}
	p.pki = p.certstore.PKI
	p.PkiOnline = true
	return nil
}


func (p *Portal) Close() {
	if p.storage != nil && p.storage.Running {
		p.storage.Close()
	}
}

func (p *Portal) ServeHttps(cfg *config.PortalHttpsConfiguration) (*config.PortalHttpsConfiguration, error) {
	if p.HttpsOnline {
		return cfg, nil
	}
	if !p.PkiOnline {
		return nil, fmt.Errorf("PKI is needed for Https Server")
	}
	//var err error
	//var tlsCert tls.Certificate
	//var interSerial *big.Int

	/*if len(cfg.Ca) < data.INDEX_SIZE {
		inter, err := p.pki.CreateIntermediate(p.S.Contact.Email, "HTTPS Intermidiate", 10, pki.DEFAULT_ASYMKEYLEN)
		if err != nil {
			return nil, fmt.Errorf("Failed to create Intermidiate with: %v", err)
		}
		interid, err := p.certstore.Add(inter)
		if err != nil {
			return nil, fmt.Errorf("Failed to Save Intermidiate to Store with: %v", err)
		}
		cfg.Ca = interid
		interSerial = inter.Serial
	} else {
		inter, err := p.certstore.Load(cfg.Ca)
		if err != nil {
			return nil, fmt.Errorf("Failed to Load Https Cert from Store with: %v", err)
		}
		interSerial = inter.Serial
	}
	if len(cfg.Cert) < data.INDEX_SIZE {


		csr := &pki.CertRequest{
			Organisation: p.S.Organisation,
			Address: p.S.Contact.Address,
			Email: p.S.Contact.Email,
			Country: p.S.Contact.Country,
			Province: p.S.Contact.County,
			Location: p.S.Contact.City,
		}
		cert, err := p.pki.GetCert(interSerial, pki.DEFAULT_ASYMKEYLEN, cfg.Address, csr)
		if err != nil {
			return nil, fmt.Errorf("Failed to create Https Certificate with: %v", err)
		}
		certid, err := p.certstore.Add(cert)
		if err != nil {
			return nil, fmt.Errorf("Failed to Save Certificate to Store with: %v", err)
		}
		cfg.Cert = certid
		tlsCert, err = cert.ToTls()
		if err != nil {
			return nil, fmt.Errorf("Failed to convert Cert to Tls Keypair with: %v", err)
		}
	} else {
		containter, err := p.certstore.Load(cfg.Cert)
		if err != nil {
			return nil, fmt.Errorf("Failed to Load Https Cert from Store with: %v", err)
		}
		tlsCert, err = containter.ToTls()
		if err != nil {
			return nil, fmt.Errorf("Failed to convert Cert to Tls Keypair with: %v", err)
		}
	}*/
	/*p.https, err = https.NewHttpsFrontend(tlsCert, cfg.ToConfig())
	if err != nil {
		return nil, fmt.Errorf("Failed to create new HttpsFrontend with: %v", err)
	}
	if cfg.HasOAuth {
		p.https.Handle("/auth", oauth.Authorize("/auth", p.authbackend, p.storage.Connect()))
	}*/
	//err = p.https.Run()
	/*if err != nil {
		return nil, fmt.Errorf("Failed to Start Https with: %v", err)
	}*/
	return cfg, nil
}

func (p *Portal) ServeHttp(cfg *HttpFrontendConfiguration) error {
	if p.StaticHttpOnline {
		return nil
	}
	var err error
	p.http, err = NewHttpFrontend(cfg)

	if err != nil {
		return fmt.Errorf("Faile to get new HttpServer with %v", err)
	}

	go func(){
		p.http.Srv.Serve(p.http.Ln)
	}()

	p.StaticHttpOnline = true
	p.http.ConnectAuthBackend(p.authbackend)

	return nil
}

func (p *Portal) createServerCert(intermediate *big.Int) (*pki.CertContainer, error){
	serverCertReq := new(pki.CertRequest)
	serverCertReq.Country = "The Stars"
	serverCertReq.Email = "info@ventose.cc"
	serverCertReq.Location = "Some Dark Place"
	serverCertReq.Organisation = "Ventose"
	serverCertReq.OrganizationalUnit = "Server?"

	cert, error := p.pki.GetCert(intermediate, 4096, "localhost", serverCertReq)

	if error != nil {
		return nil, fmt.Errorf("Can't create Cert from Intermediate with: %v", error.Error())
	}
	return cert, nil
}



func (p *Portal) Connect() (*data.StorageConnection, error) {
	if p.storage != nil && p.storage.Running {
		conn := p.storage.Connect()
		return conn, nil
	}
	return nil, fmt.Errorf("Storage not Initialised")
}

func (p *Portal) ServeStorage() error {
	if p.storage == nil {
		return fmt.Errorf("Serve Storage failed, its not initialised")
	}
	err := p.storage.Serve()
	if err == nil {
		p.StorageWaiter.Done()
		p.StorageOnline = true

	}
	return err
}

func (p *Portal) ServeAuthBackend() error {
	if !p.StorageOnline {
		p.StorageWaiter.Wait()
	}
	p.authbackend = auth.NewAuthBackend()
	if sc := p.storage.Connect(); sc != nil {
		p.authbackend.ConnectStorage(sc)
		//p.authbackend.Run()
		return nil
	} else {
		return fmt.Errorf("[Portal] Can't connect Storage in ServeAuthBackend")
	}
}

func (p *Portal) LoadStorage(cfg *data.InitialConfiguration) error {
	/*defer func() error {
		if r := recover(); r != nil {
			return fmt.Errorf("Opening Storage failed %v", r)
		}
		return nil
	}()*/
	var err error
	p.storage, err = data.OpenStorage(cfg)

	return err
}


func (p *Portal) LoadCertStore(cfgStr string) error {
	defer func() error {
		if r := recover(); r != nil {
			return fmt.Errorf("Opening Cert Store failed %v", r)
		}
		return nil
	}()

	if p.storage == nil {
		return fmt.Errorf("Load and Start Storage first!")
	}
	p.certstore = certstore.ConnectDataStorage(p.storage)
	return nil
}
