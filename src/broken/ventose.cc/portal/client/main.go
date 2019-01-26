package main

import (
	"ventose.cc/portal/serverold/ctrl"
	"crypto/rsa"
	"net/http"
	"time"
	"fmt"
	"ventose.cc/tools"
	"crypto/sha1"
	"encoding/json"
	"bytes"
	"crypto/sha256"
	"log"
	"encoding/base64"
	"crypto/rand"
	"flag"
	"os"
	"io/ioutil"
	"encoding/pem"
	"crypto/x509"
)

var keyRel string = tools.OSSP + "config" + tools.OSSP + "rsa.key"
var cfgRel string = tools.OSSP + "config" + tools.OSSP + "portal.json"
var dbRel string = tools.OSSP + "db"

var key *rsa.PrivateKey

func doCtrlServerRequest(path string, r *ctrl.Request, publicKey *rsa.PublicKey) (string, error) {
	c := new(http.Client)
	message := fmt.Sprintf("%s %s %x", r.Command, r.Date.Format(time.RFC3339), tools.GetRandomBytes(12))

	crypt, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, publicKey, []byte(message), []byte(""))

	if err != nil {
		log.Fatal(err)
	}
	r.Signature = crypt
	reqbytes, err := json.Marshal(r)

	if err != nil {
		log.Fatal(err)
	}

	buffer := new(bytes.Buffer)
	buffer.Write(reqbytes)

	cl, _ := http.NewRequest("PUT", fmt.Sprintf("http://localhost:8918%s", path), buffer)
	cl.Header.Add("Content-type", "application/json; charset=utf-8")
	hash := sha256.New()
	hashBytes := hash.Sum(reqbytes)
	b64 := base64.StdEncoding.EncodeToString(hashBytes)
	cl.Header.Add("X-Content-Hash", b64)
	//cl.Close = true
	re, err := c.Do(cl)

	if err != nil {
		panic(err.Error())
		fmt.Errorf("Panic text")
	}

	if re == nil {
		panic("Controllserver Request has no response")
	}

	if re != nil && re.StatusCode == http.StatusOK {
		defer re.Body.Close()
		d, _ := ioutil.ReadAll(re.Body)
		return string(d), nil
	}

	if re.StatusCode != http.StatusOK {
		fmt.Printf("%v\n", re)
	}

	//defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	return "", fmt.Errorf("Error")
}

func testConfigDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) || os.IsPermission(err) {
		return fmt.Errorf("Can't access %s Path, either it does not exists or you have no permission on it", path)
	}
	rsaKey := path + keyRel
	cfg := path + cfgRel

	if _, err := os.Stat(rsaKey); os.IsNotExist(err) || os.IsPermission(err) {
		return fmt.Errorf("Server Key not found!")
	}

	if _, err := os.Stat(cfg); os.IsNotExist(err) || os.IsPermission(err) {
		return fmt.Errorf("Portal Server Configuration not found.")
	}

	pemBytes, err := ioutil.ReadFile(rsaKey)

	if err != nil {
		return fmt.Errorf("Failed to Read private Key %s", err.Error())
	}

	pemBlock, _ := pem.Decode(pemBytes)

	key, err = x509.ParsePKCS1PrivateKey(pemBlock.Bytes)

	if err != nil {
		return fmt.Errorf("Failed to Parse Private Key %s", err.Error())
	}

	return nil
}

func IsCommandExist(parameter string) string {

	req := new(ctrl.Request)

	if key == nil {
		return "Key not Set"
	}
	if parameter == "start" {
		req.Command = "start"
		req.Date = time.Now()
		state, err := doCtrlServerRequest("/portal/state", req, &key.PublicKey)
		if err == nil {
			fmt.Println(state)
		} else {
			log.Fatal(err)
		}
		return parameter

	} else if parameter == "status" {
		req.Command = "status"
		req.Date = time.Now()
		state, err := doCtrlServerRequest("/portal/state", req, &key.PublicKey)
		if err == nil {
			fmt.Println(state)
		} else {
			log.Fatal(err)
		}
		return parameter

	} else if parameter == "shutdown" {
		req.Command = "shutdown"
		req.Date = time.Now()
		doCtrlServerRequest("/portal/state", req, &key.PublicKey)
		return parameter
	}

	return "Command not found"
}

func main() {

	var serverpath string
	var help, verbose bool

	flag.StringVar(&serverpath, "serverold", "not/set", "Path to Portal Server")
	flag.StringVar(&serverpath, "s", "not/set", "-serverold shortcut")
	flag.BoolVar(&help, "help", false, "Display help")
	flag.BoolVar(&help, "h", false, "Display help")
	flag.BoolVar(&verbose, "verbose", false, "Talk a lot!")
	flag.BoolVar(&verbose, "v", false, "Talk a lot!")


	flag.Parse()
	params := flag.Args()

	if help {
		fmt.Println("Usage: portalctrl -serverold=<path> command")

		flag.PrintDefaults()

		fmt.Println("available Commands are:")
		fmt.Println("\tshutdown")
		return
	}

	err := testConfigDir(serverpath)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	switch params[0] {
	case IsCommandExist(params[0]): {
		fmt.Printf("%s processed\n", params[0])
	}
	default:
		fmt.Printf("%s is no kown Command", params[0])
	}



}
