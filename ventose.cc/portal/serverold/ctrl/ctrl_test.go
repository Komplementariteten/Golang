package ctrl

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
	"ventose.cc/data"
	"ventose.cc/tools"
)

func TestStartControllServer(t *testing.T) {
	cfg := new(ControllServerConfiguration)
	cfg.ServerKeyPath = "/tmp/"
	cfg.ListenPort = 12345
	cfg.CtrlServerLogFile = "/tmp/crtlog.log"
	//StartControllServer(cfg)
}

func getPublicKey(t *testing.T, cfg *ControllServerConfiguration) *rsa.PublicKey {

	rsaFile, err := os.Open(fmt.Sprintf("%s%srsa.key", cfg.ServerKeyPath, tools.OSSP))
	buffer := new(bytes.Buffer)
	n, err := buffer.ReadFrom(rsaFile)

	if err != nil {
		t.Fatal(err)
	}

	if n == 0 {
		t.Fatal("No Data read from RSA Key file")
	}

	keyBytes, _ := pem.Decode(buffer.Bytes())

	key, err := x509.ParsePKCS1PrivateKey(keyBytes.Bytes)

	return &key.PublicKey
}

func doRequest(t *testing.T, path string, r *Request, publicKey *rsa.PublicKey) *http.Response {
	time.Sleep(2 * time.Second)
	c := new(http.Client)
	message := fmt.Sprintf("%s %s %x", r.Command, r.Date.Format(time.RFC3339), tools.GetRandomBytes(12))

	crypt, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, publicKey, []byte(message), []byte(""))

	if err != nil {
		t.Fatal(err)
	}
	r.Signature = crypt
	reqbytes, err := json.Marshal(r)

	if err != nil {
		t.Fatal(err)
	}

	buffer := new(bytes.Buffer)
	buffer.Write(reqbytes)

	cl, _ := http.NewRequest("PUT", fmt.Sprintf("http://localhost:12345%s", path), buffer)
	cl.Header.Add("Content-type", "application/json; charset=utf-8")
	hash := sha256.New()
	hashBytes := hash.Sum(reqbytes)
	b64 := base64.StdEncoding.EncodeToString(hashBytes)
	cl.Header.Add("X-Content-Hash", b64)

	resp, _ := c.Do(cl)
	//defer resp.Body.Close()
	return resp
}

func evaluadeHttpResponse(t *testing.T, resp *http.Response) {
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("Request failed with: %d %s", resp.StatusCode, bodyBytes)
	}
}

func startServer() (*ControllServerConfiguration, chan bool) {
	cfg := new(ControllServerConfiguration)
	cfg.ServerKeyPath = "/tmp/"
	cfg.ListenPort = 12345
	cfg.CtrlServerLogFile = "/tmp/crtlog.log"
	waiter := StartControllServer(cfg)
	return cfg, waiter
}

func stopServer(t *testing.T, pubKey *rsa.PublicKey) {
	req := new(Request)
	req.Command = "shutdown"
	req.Date = time.Now()
	resp := doRequest(t, "/portal/state", req, pubKey)
	evaluadeHttpResponse(t, resp)
}

func TestAppHandle_SetupStorage(t *testing.T) {
	cfg, waiter := startServer()
	publicKey := getPublicKey(t, cfg)
	req := new(Request)
	req.Command = "setup"
	req.Date = time.Now()
	req.Parameter = make(map[string]string)

	storageConfig := new(data.InitialConfiguration)
	storageConfig.AuthKey = tools.GetRandomAsciiString(10)
	storageConfig.ConfigPath = "/tmp/ledis.cfg"
	storageConfig.DbPath = "/tmp/data"
	storageConfig.MaxDatabases = 12
	storageConfig.PathAccessMode = 0770

	jsonBytes, err := json.Marshal(storageConfig)

	if err != nil {
		t.Fatalf("Can't Serialize storageConfig: %v", err)
	}

	json := base64.StdEncoding.EncodeToString(jsonBytes)

	req.Parameter["StorageSetup"] = json
	resp := doRequest(t, "/portal/state", req, publicKey)
	evaluadeHttpResponse(t, resp)

	if _, err := os.Stat(storageConfig.ConfigPath); os.IsNotExist(err) {
		t.Fatalf("Leids Configuration File not found: %v", err)
	}

	if _, err := os.Stat(storageConfig.DbPath); os.IsNotExist(err) {
		t.Fatalf("Leids Data Directory not found: %v", err)
	}

	stopServer(t, publicKey)
	<-waiter
}

func TestAppHandle_ServeHTTP(t *testing.T) {
	cfg, waiter := startServer()
	publicKey := getPublicKey(t, cfg)

	req := new(Request)
	req.Command = "help"
	req.Date = time.Now()
	resp := doRequest(t, "/portal/state", req, publicKey)
	//evaluadeHttpResponse(t, resp)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	var helpMap map[string]string
	json.Unmarshal(bodyBytes, &helpMap)

	for key := range helpMap {
		if key != "shutdown" && key != "help" {
			req.Command = key
			req.Date = time.Now()
			resp = doRequest(t, "/portal/state", req, publicKey)
			evaluadeHttpResponse(t, resp)
		}
	}

	req.Command = "shutdown"
	req.Date = time.Now()
	resp = doRequest(t, "/portal/state", req, publicKey)
	evaluadeHttpResponse(t, resp)

	<-waiter
}
