package pki

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"log"
	"os"
	"path"
)

const pki_path = "sec"
const key_file = "key.pem"

var prvKey *ecdsa.PrivateKey
var sep = []byte{0xFE, 0xED, 0xA0}

func Init() {
	if prvKey != nil {
		return
	}

	if _, err := os.Stat(pki_path); os.IsNotExist(err) {
		os.Mkdir(pki_path, 0700)
	}

	if _, err := os.Stat(path.Join(pki_path, key_file)); os.IsNotExist(err) {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			log.Fatalln(err)
		}
		prvKey = key
		key_bytes, key_err := x509.MarshalECPrivateKey(prvKey)
		if key_err != nil {
			log.Fatalln(key_err)
		}
		werr := os.WriteFile(path.Join(pki_path, key_file), key_bytes, 0600)
		if werr != nil {
			log.Fatalln(werr)
		}
	}

	pbytes, perr := os.ReadFile(path.Join(pki_path, key_file))
	if perr != nil {
		log.Fatalln(perr)
	}
	key, kerr := x509.ParseECPrivateKey(pbytes)
	if kerr != nil {
		log.Fatalln(kerr)
	}
	prvKey = key
}

func Sign(bytes []byte) ([]byte, error) {
	result := make([]byte, 0)
	sig, sig_err := ecdsa.SignASN1(rand.Reader, prvKey, bytes)
	if sig_err != nil {
		return nil, sig_err
	}
	result = append(result, bytes...)
	rnd := make([]byte, 3)
	_, err := rand.Read(rnd)
	if err != nil {
		return nil, err
	}
	result = append(result, rnd...)
	result = append(result, sep...)
	_, err = rand.Read(rnd)
	if err != nil {
		return nil, err
	}
	result = append(result, rnd...)
	result = append(result, sig...)
	return result, nil
}

func Verify(b []byte) ([]byte, error) {
	parts := bytes.Split(b, sep)
	if len(parts) != 2 {
		return nil, errors.New("invalid signature")
	}
	payload := parts[0][:len(parts[0])-3]
	sig := parts[1][3:len(parts[1])]

	if ecdsa.VerifyASN1(&prvKey.PublicKey, payload, sig) {
		return payload, nil
	}
	return nil, errors.New("invalid signature")
}
