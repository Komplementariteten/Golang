package pki

import (
	"crypto/x509"
	"ventose.cc/data"
	"ventose.cc/pki/export"
	"crypto/x509/pkix"
	"fmt"
)

type CertRequest struct {
	Email              string
	Country            string
	Location           string
	Organisation       string
	OrganizationalUnit string
	Province           string
	Address            string
}

type CA struct {
	Organisation 		string
	DefaultCertValidYears	int
	Id			[]byte
	Cert			*export.Cert
	Key			*export.Key
	IntermediatePool	*x509.CertPool
	Revoked			*pkix.CertificateList
	storage			*data.StorageConnection
}

//Private Methods
func loadCACert(id []byte, con *data.StorageConnection) (*export.Cert, error){
	req := new(data.StorageRequest)
	req.Type = data.ReadRequest
	req.Element = id
	req.Content = new(export.Cert)
	con.RequestChannel <- req
	resp := <- con.ResponseChannel
	if resp.Error != nil {
		return nil, fmt.Errorf("Failed to load CA Cert with: %v", resp.Error)
	}
	if cert, ok := resp.Content.(*export.Cert); ok {
		return cert, nil
	}
	return nil, fmt.Errorf("Failed to Convert CA Cert Request to export.Cert")
}

func loadCAExport(id []byte, con *data.StorageConnection) (*export.CAExport, error){
	req := new(data.StorageRequest)
	req.Type = data.ReadRequest
	req.Element = id
	req.Content = new(export.CAExport)
	con.RequestChannel <- req
	resp := <- con.ResponseChannel
	if resp.Error != nil {
		return nil, fmt.Errorf("Failed to load CA Export with: %v", resp.Error)
	}
	if cert, ok := resp.Content.(*export.CAExport); ok {
		return cert, nil
	}
	return nil, fmt.Errorf("Failed to Convert CA Cert Request to export.CAExport")

}

func loadKeyExport(id []byte, con *data.StorageConnection) (*export.Key, error){
	req := new(data.StorageRequest)
	req.Type = data.ReadRequest
	req.Element = id
	req.Content = new(export.Key)
	con.RequestChannel <- req
	resp := <- con.ResponseChannel
	if resp.Error != nil {
		return nil, fmt.Errorf("Failed to load Key Export with: %v", resp.Error)
	}
	if cert, ok := resp.Content.(*export.Key); ok {
		return cert, nil
	}
	return nil, fmt.Errorf("Failed to Convert Key Request to export.Key")
}

func loadRevokedExport(id []byte, con *data.StorageConnection) (*export.Revoked, error) {

}

// General Methods
func LoadFromStorage(id []byte, con *data.StorageConnection) (*CA, error) {

	ca := new(CA)

	export, exerr := loadCAExport(id, con)
	if exerr != nil {
		return nil, exerr
	}
	ca.Id = export.Id
	ca.DefaultCertValidYears = export.DefaultValidity

	cert, exerr := loadCACert(export.Cert, con)
	if exerr != nil {
		return nil, exerr
	}
	ca.Cert = cert
	ca.Organisation = cert.Data.Subject.Organization
	key, exerr := loadKeyExport(export.Key, con)
	if exerr != nil {
		return nil, exerr
	}
	ca.Key = key


	return ca, nil
}

func (ca *CA) ConnectStorage(con *data.StorageConnection) {
	ca.storage = con
}


