package export

import (
	"ventose.cc/data"
	"crypto/x509"
	"crypto"
	"crypto/x509/pkix"
)


type Revoked struct {
	data.RequestContent
	Id []byte
	List pkix.CertificateList
}

type CAExport struct {
	data.RequestContent
	Id 		[]byte
	Cert 		[]byte
	Key		[]byte
	Revoked 	[]byte
	Intermediate	[]byte
	DefaultValidity int
}

type Intermediate struct {
	Cert
	Organisation string
	Parent       []byte
	Id           []byte
}

type Cert struct {
	data.RequestContent
	Data   x509.Certificate
	KeyId  []byte
	Id     []byte
	Serial string `search:"cert_serial"`
}

type Key struct {
	data.RequestContent
	Data       crypto.PrivateKey
	CertSerial string `search:"key_certserial"`
	Id         []byte
}