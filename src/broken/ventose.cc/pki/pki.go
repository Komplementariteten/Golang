package pki

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/gob"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"time"
	"ventose.cc/tools"
	"crypto/tls"
)

const (
	DEFAULT_CERT_VALIDITY = 200
	DEFAULT_ASYMKEYLEN    = 6168
	DEFAULT_SYMKEYLEN     = 256
	CERT_REVOKED          = "Cert has been Revoked"
)


type PKIExport struct {
	Crt                   []byte
	Organisation          string
	DefaultCertValidYears int
	Key                   []byte
	Serial                big.Int
}

type CertContainer struct {
	Serial     *big.Int
	Crt        []byte
	Key        *rsa.PrivateKey
	ValidUntil time.Time
}

type PKI struct {
	storageKey            []byte
	rootCrt               *x509.Certificate
	Serial                *big.Int
	Organisation          string
	DefaultCertValidYears int
	key                   *rsa.PrivateKey
	Intermediate          map[string]*CertContainer
	IntermediatePool      *x509.CertPool
	Revoked               []pkix.RevokedCertificate
	Crl                   []byte
	CrlTTL                time.Time
}

/*
CertContainer Type
 */

func (c *CertContainer) ToTls() (tls.Certificate, error) {


	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Crt})
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: c.PlainKey()})
	return tls.X509KeyPair(certPem, keyPem)
}

func (c *CertContainer) LoadCert(pemBytes []byte, passPhrase []byte) error {

	pem, _ := pem.Decode(pemBytes)

	if len(pem.Bytes) == 0 {
		return fmt.Errorf("Can't pem Decode Cert")
	}

	if x509.IsEncryptedPEMBlock(pem) {
		plain, e := x509.DecryptPEMBlock(pem, passPhrase)
		if e != nil {
			return fmt.Errorf("Can't Cert DecryptPEMBlock: %v", e)
		}
		c.Crt = plain
	}
	return nil
}

func (c *CertContainer) ExportCert(passPhrase []byte) []byte {

	pemData, err := x509.EncryptPEMBlock(rand.Reader, "CERTIFICATE", c.Crt, passPhrase, x509.PEMCipherAES256)
	if err != nil {
		panic(err)
	}
	encdata := pem.EncodeToMemory(pemData)

	return encdata
}

func (c *CertContainer) LoadKey(pemBytes []byte, passPhrase []byte) error {

	pem, _ := pem.Decode(pemBytes)

	if len(pem.Bytes) == 0 {
		return fmt.Errorf("Can't pem Decode Key")
	}

	if x509.IsEncryptedPEMBlock(pem) {
		plain, e := x509.DecryptPEMBlock(pem, passPhrase)
		if e != nil {
			return fmt.Errorf("Can't Key DecryptPEMBlock: %v", e)
		}

		key, err := x509.ParsePKCS1PrivateKey(plain)

		if err != nil {
			return fmt.Errorf("Can't Parse Private Key %v", err)
		}

		c.Key = key
	}
	return nil
}

func (c *CertContainer) PlainKey() []byte {
	m := x509.MarshalPKCS1PrivateKey(c.Key)
	return m
}

func (c *CertContainer) ExportKey(passPhrase []byte) []byte {
	marshal := x509.MarshalPKCS1PrivateKey(c.Key)

	pemBlock, err := x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", marshal, passPhrase, x509.PEMCipherAES256)
	if err != nil {
		return nil
	}
	encrypted := pem.EncodeToMemory(pemBlock)
	return encrypted
}

func (c *CertContainer) GetX509Cert() (*x509.Certificate, error) {
	cert, error := x509.ParseCertificate(c.Crt)
	if error != nil {
		return nil, error
	}
	return cert, nil
}

/*
PKI Type
 */

func (p *PKI) IsRevoked(cert *x509.Certificate) bool {
	for i := 0; i < len(p.Revoked); i++ {
		if p.Revoked[i].SerialNumber == cert.SerialNumber {
			return true
		}
	}
	return false
}

func (p *PKI) CreateCRL(ttl time.Time) (crlBytes []byte, err error) {
	crlBytes, err = p.rootCrt.CreateCRL(rand.Reader, p.key, p.Revoked, time.Now(), ttl)
	return
}

func (p *PKI) Revoke(cert *x509.Certificate) {
	revoke := new(pkix.RevokedCertificate)
	revoke.RevocationTime = time.Now()
	revoke.SerialNumber = cert.SerialNumber
	revoke.Extensions = cert.Extensions
	p.Revoked = append(p.Revoked, *revoke)
	var err error
	p.Crl, err = p.CreateCRL(p.CrlTTL)
	if err != nil {
		panic(err)
	}
}

func (p *PKI) Verify(cert *x509.Certificate, opts x509.VerifyOptions) error {

	if p.IsRevoked(cert) {
		return fmt.Errorf(CERT_REVOKED)
	}

	opts.Intermediates = p.IntermediatePool
	opts.Roots = x509.NewCertPool()

	_, error := cert.Verify(opts)

	if error != nil {
		return error
	}
	return nil
}

func (p *PKI) GetCert(parent *big.Int, KeySize int, cn string, req *CertRequest) (*CertContainer, error) {
	if _, ok := p.Intermediate[parent.String()]; !ok {
		return nil, fmt.Errorf("Parent Intermediate Certifikate not found")
	}

	template := getCertTamplate()

	template.DNSNames = []string{cn}
	template.EmailAddresses = []string{req.Email}
	template.Subject = pkix.Name{
		CommonName: cn,
	}
	if req.Country != "" {
		template.Subject.Country = []string{req.Country}
	}

	if req.Location != "" {
		template.Subject.Locality = []string{req.Location}
	}

	if req.Organisation != "" {
		template.Subject.Organization = []string{req.Organisation}
	} else {
		template.Subject.Organization = []string{p.Organisation}
	}

	if req.Email != "" {
		template.EmailAddresses = []string{req.Email}
	}

	if req.OrganizationalUnit != "" {
		template.Subject.OrganizationalUnit = []string{req.OrganizationalUnit}
	}

	if req.Province != "" {
		template.Subject.Province = []string{req.Province}
	}

	if req.Address != "" {
		template.Subject.StreetAddress = []string{req.Address}
	}

	kp, err := rsa.GenerateKey(rand.Reader, KeySize)
	if err != nil {
		return nil, fmt.Errorf("Error in generating Private Key: %v", err)
	}

	parentContainer := p.Intermediate[parent.String()]

	parentCrt, err := x509.ParseCertificate(parentContainer.Crt)
	if err != nil {
		return nil, fmt.Errorf("Parsing Parent Cert failed with: %v", err)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, parentCrt, kp.Public(), parentContainer.Key)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Cert with: %v", err)
	}

	cert := new(CertContainer)
	cert.Crt = derBytes
	cert.Key = kp
	cert.ValidUntil = time.Now().AddDate(p.DefaultCertValidYears, 0, 0)

	return cert, nil
}

func (p *PKI) AddIntermediateFromExport(export []byte, key *rsa.PrivateKey) error {

	pemBlock, _ := pem.Decode(export)

	if x509.IsEncryptedPEMBlock(pemBlock) {
		container := new(CertContainer)
		plain, e := x509.DecryptPEMBlock(pemBlock, p.storageKey)
		if e != nil {
			return fmt.Errorf("Can't add Intermediate to PKI DecryptPEMBlock failed with: %v", e)
		}
		cert, err := x509.ParseCertificate(plain)

		if err != nil {
			return fmt.Errorf("Can't Parse Intermediate Cert with: %v", err)
		}

		err = cert.CheckSignatureFrom(p.rootCrt)

		if err != nil {
			return fmt.Errorf("Can't validate Signature on Intermediate Cert with: %v", err)
		}

		container.Crt = plain
		container.ValidUntil = cert.NotAfter
		container.Serial = cert.SerialNumber
		container.Key = key
		err = p.AddIntermediate(container)

		if err != nil {
			fmt.Errorf("Can't add Intermediate Cert with: %v", err)
		}

	} else {
		return fmt.Errorf("Can't add Intermediate to PKI Pem Block is not Encrypted")
	}

	return nil
}

func (p *PKI) AddIntermediate(c *CertContainer) error {
	serial := c.Serial.String()

	if _, ok := p.Intermediate[serial]; ok {
		return fmt.Errorf("PKI already contains a Intermediate with this Serial")
	}

	p.Intermediate[serial] = c
	p.IntermediatePool.AppendCertsFromPEM(c.Crt)
	return nil
}

func (p *PKI) CreateIntermediate(EmailAddress string, Title string, ValidYears int, KeySize int) (*CertContainer, error) {

	cert := new(CertContainer)
	template := getRootCaTamplate(p.Organisation)
	template.EmailAddresses = []string{EmailAddress}
	template.NotBefore = time.Now()
	template.NotAfter = time.Now().AddDate(ValidYears, 0, 0)
	template.Issuer = pkix.Name{
		Organization: []string{p.Organisation},
		SerialNumber: p.Serial.String(),
	}
	template.Subject = pkix.Name{
		Organization: []string{p.Organisation},
		CommonName:   fmt.Sprint(p.Organisation, " ", Title),
	}
	kp, err := rsa.GenerateKey(rand.Reader, KeySize)
	if err != nil {
		return nil, fmt.Errorf("Can't create RSA Key: %v", err)
	}
	cert.Key = kp

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, p.rootCrt, kp.Public(), p.key)

	if err != nil {
		return nil, fmt.Errorf("Can't create Imtermedate Cert %v", err)
	}

	cert.Crt = derBytes
	cert.Serial = template.SerialNumber
	cert.ValidUntil = template.NotAfter
	p.Intermediate[cert.Serial.String()] = cert
	p.IntermediatePool.AppendCertsFromPEM(derBytes)
	return cert, nil
}

func (p *PKI) ValidateByCaRaw(derBytes []byte) (bool, error) {

	cert, e := x509.ParseCertificate(derBytes)

	e = cert.CheckSignatureFrom(p.rootCrt)
	if e != nil {
		return false, fmt.Errorf("Cert is not Signed by Root CA: %v", e)
	}
	return true, nil
}

func (p *PKI) Export() ([]byte, error) {

	export := new(PKIExport)
	export.DefaultCertValidYears = p.DefaultCertValidYears
	export.Crt = p.rootCrt.Raw
	export.Organisation = p.Organisation
	export.Key = x509.MarshalPKCS1PrivateKey(p.key)
	export.Serial = *p.Serial

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(export)
	if err != nil {
		return nil, fmt.Errorf("Gob Encode failed with: %v", err)
	}
	pemBlock, err := x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", buffer.Bytes(), p.storageKey, x509.PEMCipherAES256)
	if err != nil {
		return nil, err
	}
	encrypted := pem.EncodeToMemory(pemBlock)
	return encrypted, nil
}

/**
Create base Data
*/
func NewPki(organisation string, secure []byte, KeySize int) (root *PKI, e error) {

	template := getRootCaTamplate(organisation)

	kp, err := rsa.GenerateKey(rand.Reader, KeySize)

	if err != nil {
		return nil, fmt.Errorf("Can't Create RSA Key:%v", err)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &kp.PublicKey, kp)

	if err != nil {
		return nil, fmt.Errorf("Can't create Root CA Cert with: %v", err)
	}

	root = new(PKI)
	root.key = kp
	root.rootCrt, err = x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, fmt.Errorf("Can't Parse Certificate with: %v", err)
	}
	root.Organisation = organisation
	root.DefaultCertValidYears = DEFAULT_CERT_VALIDITY
	root.storageKey = secure
	root.Serial = tools.GetSerial()
	root.Intermediate = make(map[string]*CertContainer)
	root.IntermediatePool = x509.NewCertPool()

	return root, nil
}

func LoadPkiFromPEMBlock(pemEncoded []byte, key []byte) (root *PKI, e error) {
	pemBlock, rest := pem.Decode(pemEncoded)
	if len(rest) > 0 {
		return nil, fmt.Errorf("Pem Data containse Noise")
	}
	var plain []byte
	if x509.IsEncryptedPEMBlock(pemBlock) {
		plain, e = x509.DecryptPEMBlock(pemBlock, key)
		if e != nil {
			return nil, fmt.Errorf("DecryptPEMBlock failed with: %v", e)
		}
	} else {
		return nil, fmt.Errorf("PEm Block is not Encrypted")
	}

	export := new(PKIExport)
	var buffer bytes.Buffer
	buffer.Write(plain)
	decoder := gob.NewDecoder(&buffer)
	e = decoder.Decode(export)
	if e != nil {
		return nil, fmt.Errorf("Gob Decode failed with: %v", e)
	}

	root = new(PKI)
	root.Organisation = export.Organisation
	root.storageKey = key
	root.DefaultCertValidYears = export.DefaultCertValidYears
	root.key, e = x509.ParsePKCS1PrivateKey(export.Key)
	root.rootCrt, e = x509.ParseCertificate(export.Crt)
	if e != nil {
		return nil, fmt.Errorf("Can't Load PEM Block: %v", e)
	}
	root.Serial = &export.Serial

	return root, nil
}

func getCertTamplate() x509.Certificate {
	keyID := make([]byte, DEFAULT_ASYMKEYLEN)
	_, err := rand.Read(keyID)

	if err != nil {
		log.Fatalf("Cant get Root CA Cert Template KeyId with %s", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(DEFAULT_CERT_VALIDITY, 0, 0)
	certTemplate := x509.Certificate{
		SerialNumber:          tools.GetSerial(),
		Subject:               pkix.Name{},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		SubjectKeyId:          keyID,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment | x509.KeyUsageContentCommitment | x509.KeyUsageKeyAgreement,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		BasicConstraintsValid: true,
		IsCA:                        false,
		SignatureAlgorithm:          x509.SHA256WithRSA,
		PermittedDNSDomainsCritical: true,
		/*DNSNames: []string{"*.ventose.cc", "www.ventose.cc","ventose.cc"},
		PermittedDNSDomainsCritical: true,
		PermittedDNSDomains: []string { "ventose.cc"},*/
	}

	return certTemplate
}

/*
Root CA x509 Template
*/

func getRootCaTamplate(organisationName string) x509.Certificate {

	keyID := make([]byte, DEFAULT_ASYMKEYLEN)
	_, err := rand.Read(keyID)

	if err != nil {
		log.Fatalf("Cant get Root CA Cert Template KeyId with %s", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(DEFAULT_CERT_VALIDITY, 0, 0)
	rootCaTemplate := x509.Certificate{
		SerialNumber: tools.GetSerial(),
		Subject: pkix.Name{
			Organization: []string{organisationName},
			CommonName:   fmt.Sprint(organisationName, " ROOT CA"),
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		SubjectKeyId:          keyID,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:               true,
		SignatureAlgorithm: x509.SHA256WithRSA,
		MaxPathLen:         5,
		/*DNSNames: []string{"*.ventose.cc", "www.ventose.cc","ventose.cc"},
		PermittedDNSDomainsCritical: true,
		PermittedDNSDomains: []string { "ventose.cc"},*/
	}

	return rootCaTemplate
}
