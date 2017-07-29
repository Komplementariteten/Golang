package export

import (
	"os"
	"ventose.cc/data"
	"testing"
	"crypto/x509"
)

const (
	DbCfg  = "/tmp/ledif.cfg"
	DbPath = "/tmp/testdb"
)

func GetTestCfg() *data.InitialConfiguration {
	cfg := new(data.InitialConfiguration)
	cfg.ConfigPath = DbCfg
	cfg.AuthKey = "12345"
	cfg.DbPath = DbPath
	cfg.MaxDatabases = 10
	cfg.PathAccessMode = os.FileMode(0700)
	return cfg
}

func TestTest(t *testing.T) {
	t.Log("test")
}



func TestExportCert(t *testing.T){

	ccfg := GetTestCfg()
	d, err := data.OpenStorage(ccfg)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	d.Serve()
	con := d.Connect()
	defer con.Close()

	cert := &x509.Certificate{
		EmailAddresses: []string{"osiegemund@gmail.com"},
		IsCA: true,
	}
	c := new(Cert)
	c.Data = *cert

	req := new(data.StorageRequest)
	req.Content = c
	req.Type = data.CreateRequest

	con.RequestChannel <- *req
	resp := <- con.ResponseChannel

	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	req.Content = new(Cert)
	req.Element = resp.Affected

	req.Type = data.ReadRequest

	con.RequestChannel <- *req
	resp = <- con.ResponseChannel

	if resp.Error != nil {
		t.Fatalf("Failed to Read Cert %v", resp.Error)
	}


	if cer, ok := resp.Content.(*Cert); !ok && !cer.Data.IsCA {
		t.Fatal("Failed to retrive correct Cert %v:%v", ok, resp.Content.(*Cert))
	}

}