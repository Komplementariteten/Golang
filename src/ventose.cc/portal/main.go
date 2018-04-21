package main

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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"ventose.cc/portal/server/ctrl"
	"ventose.cc/tools"
)

const DefaultConfigPath = "/etc/portal.json"

func getNewRsaKey(path string, size uint) (*rsa.PrivateKey, error) {

	key, err := rsa.GenerateKey(rand.Reader, int(size))

	if err != nil {
		return nil, fmt.Errorf("Controll Server Key generation Failed with %s", err)
	}

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	ioutil.WriteFile(path+tools.OSSP+"rsa.key", pemData, 0600)
	return key, nil
}

func doCtrlServerRequest(path string, r *ctrl.Request, publicKey *rsa.PublicKey) *http.Response {
	time.Sleep(2 * time.Second)
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

	resp, _ := c.Do(cl)
	//defer resp.Body.Close()
	return resp
}

func main() {

	var configPtr, keypath, defaultPath string
	var init, testrun, verbose, force bool
	var ctlKeySizse uint
	/*configPtr := flag.String("config", DefaultConfigPath, "Path to Portal-Server Configuration")
	keypath := flag.String("keyPath", "", "")*/

	// Boolean Options
	flag.BoolVar(&init, "init", false, "Generate a new default Configuration")
	flag.BoolVar(&testrun, "test", false, "Just a Teststart")
	flag.BoolVar(&verbose, "verbose", false, "Bee verbosy")
	flag.BoolVar(&force, "force", false, "Force Questions with y")

	// String Options
	flag.StringVar(&configPtr, "config", DefaultConfigPath, "Path to Portal-Server Configuration")
	flag.StringVar(&keypath, "key-path", "", "Directory to include RSA Key")
	flag.StringVar(&defaultPath, "path", "/def/ault/path", "Path to Server Data and Config")

	// Int Options
	flag.UintVar(&ctlKeySizse, "keysize", 4096, "Size of the Controll Server private Key")

	flag.Parse()

	ctrlCfg := new(ctrl.ControllServerConfiguration)
	var pubKey *rsa.PublicKey

	if init {
		if verbose {
			fmt.Println("Initialize Portal Server Configuration")
		}
		ctrlCfg, _ = ctrl.GetControllServerDefaultConfig()
		if _, err := os.Stat(defaultPath); len(defaultPath) > 0 && !os.IsNotExist(err) {
			if verbose {
				fmt.Println("Portal Configuration Directory found")
			}

			fis, _ := ioutil.ReadDir(defaultPath)

			if len(fis) > 2 {
				if verbose {
					if !force {
						fmt.Println("Directory " + defaultPath + " not empty")

						fmt.Printf("Should we delete anything in it? [y/n]\n")
						test := tools.TestStringInput("yn")

						if strings.EqualFold(test, "n") {
							fmt.Println("No Server initialisation, exiting")
							return
						}
					}

				} else if !force {
					log.Fatalln("Initialisation failed, Directory " + defaultPath + " not empty")
				}

				os.RemoveAll(defaultPath)
				os.Mkdir(defaultPath, 0700)
			}

			ctrlCfg.SetBasePath(defaultPath)
		}
		key, err := getNewRsaKey(ctrlCfg.ServerKeyPath, ctlKeySizse)

		if err != nil {
			log.Fatalln("Error initializing Portal Server, failed to generate RSA Key")
		}
		pubKey = &key.PublicKey

		err = tools.DumpConfig(ctrlCfg.CtrlServerConfig, ctrlCfg)

		if err != nil {
			log.Fatal(err)
		}

	} else {
		if verbose {
			fmt.Printf("Loading Server Configuration from %s\n", defaultPath)
		}
		cfgPath := defaultPath + tools.OSSP + "config" + tools.OSSP + "portal.json"
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) || os.IsPermission(err) {
			log.Fatalf("Can't load Config from %s", cfgPath)
		} else {
			err := tools.LoadFromJsonFile(cfgPath, ctrlCfg)
			if err != nil {
				log.Fatalf("Configuration File invalid!\n%v", err)
			}
		}
	}
	waiter := make(chan bool)
	go func() {
		if verbose {
			fmt.Println("Starting Portal Controll Server")
		}
		waiter = ctrl.StartControllServer(ctrlCfg)
	}()

	time.Sleep(8 * time.Second)

	if testrun {
		if verbose {
			fmt.Println("Try to shutdown Portal Controll Server")
		}
		r := new(ctrl.Request)
		r.Command = "shutdown"
		r.Date = time.Now()

		resp := doCtrlServerRequest("/portal/state", r, pubKey)

		if resp == nil {
			panic("No Response")
		}

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			log.Fatalf("Failed to Shutdown Server on test Flag with: %d %s\n", resp.StatusCode, bodyBytes)
		}

	} else {
		if verbose {
			fmt.Println("Controll Server Configuration seems ok")
		} else {
			fmt.Println("OK")
		}
	}

	<-waiter
	if verbose {
		fmt.Println("Shutdown complete")
	}
}
