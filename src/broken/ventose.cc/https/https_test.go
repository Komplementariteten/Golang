package https

import (
	"time"
	"ventose.cc/tools"
	"crypto/x509"
	"crypto/rsa"
	"crypto/x509/pkix"
	"crypto/rand"
	"crypto/tls"
	"testing"
	"net/http"
	"fmt"
	"encoding/pem"
	"os"
	"io/ioutil"
)

func newCert(t *testing.T) tls.Certificate{
	req := new(x509.Certificate)
	req.EmailAddresses = []string{ "info@ventose.cc"}
	req.DNSNames = []string{"localhost"}
	req.Subject = pkix.Name{
		CommonName: "localhost",
	}

	req.NotBefore= time.Now()
	req.NotAfter = req.NotBefore.AddDate(0, 0, 1)
	req.SerialNumber = tools.GetSerial()
	req.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment | x509.KeyUsageContentCommitment | x509.KeyUsageKeyAgreement
	req.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageAny}
	req.IsCA = false
	req.SignatureAlgorithm = x509.SHA384WithRSA

	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pkcs1 := x509.MarshalPKCS1PrivateKey(rsaKey)

	derBytes, err := x509.CreateCertificate(rand.Reader, req, req, rsaKey.Public(), rsaKey)

	if err != nil {
		t.Fatalf("Failed to Create Cert with: %v",err)
	}


	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: pkcs1})

	cer, error := tls.X509KeyPair(pemCert, pemKey)
	if error != nil {
		t.Fatalf("Failed to Load KeyPair with: %v", error)
	}
	return cer
}

func newHttpsConfig(t *testing.T) *HttpsConfiguration {
	c := new(HttpsConfiguration)
	c.Address = ":12345"
	return c
}

func TestNewHttpsFrontend(t *testing.T) {
	c := newCert(t)
	cfg := newHttpsConfig(t)
	_, e := NewHttpsFrontend(c,cfg)
	if e != nil {
		t.Fatal(e)
	}
}
func testFail(url string,t *testing.T) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify:true,
		},
	}
	cl := &http.Client{
		Transport: tr,
	}

	_, err := cl.Get(url)
	if err == nil {
		t.Fatal("Url found but should fail!")
	}

}

func testUrl(url string,t *testing.T) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify:true,
		},
	}
	cl := &http.Client{
		Transport: tr,
	}

	r, err := cl.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	if r.StatusCode != 200 {
		t.Fatalf("Get Returned %d %s for %s", r.StatusCode, r.Status, url)
	}

}

func TestHttpsFrontend_Handle(t *testing.T) {
	c := newCert(t)
	cfg := newHttpsConfig(t)
	f, e := NewHttpsFrontend(c,cfg)
	if e != nil {
		t.Fatal(e)
	}
	f.Handle("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	})
	e = f.Run()
	defer f.ln.Close()
	if e != nil {
		t.Fatal(e)
	}
	testUrl("https://localhost" + cfg.Address + "/hello", t)
}
func TestHttpsFrontend_Handle2(t *testing.T) {
	c := newCert(t)
	cfg := newHttpsConfig(t)
	f, e := NewHttpsFrontend(c,cfg)
	if e != nil {
		t.Fatal(e)
	}
	f.Handle("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	})
	e = f.Run()
	defer f.ln.Close()
	if e != nil {
		t.Fatal(e)
	}
	testUrl("https://localhost" + cfg.Address + "/hello", t)
	f.Handle("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	})
	testUrl("https://localhost" + cfg.Address + "/test", t)

}
func TestHttpsFrontend_Static(t *testing.T) {

	os.Mkdir("/tmp/httpTest", 0777)

	err := ioutil.WriteFile("/tmp/httpTest/t.o", []byte("test"), 0666)
	if err != nil {
		t.Fatal(err)
	}

	c := newCert(t)
	cfg := newHttpsConfig(t)
	cfg.StaticPath = "/tmp/httpTest"
	f, e := NewHttpsFrontend(c,cfg)
	if e != nil {
		t.Fatal(e)
	}
	f.Handle("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	})
	e = f.Run()
	defer f.ln.Close()
	if e != nil {
		t.Fatal(e)
	}
	testUrl("https://localhost" + cfg.Address + "/hello", t)

	testUrl("https://localhost" + cfg.Address + "/resource/t.o", t)

	testUrl("https://localhost" + cfg.Address + "/resource/t.o2", t)
}