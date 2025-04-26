package pki

import (
	"FileBrowser/env"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"log"
	"os"
	"path"
)

const pki_path = "sec"
const key_file = "key.pem"
const aes_file = "key2.pem"

var prvKey *ecdsa.PrivateKey
var encKey []byte
var sep = []byte{0xFE, 0xED, 0xA0}

func Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	return append(nonce, ciphertext...), nil
}

func Decrypt(ciphertext []byte) ([]byte, error) {

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func Init() {
	if prvKey != nil {
		return
	}
	pki_base := path.Join(env.Enviroment().BaseDir, pki_path)
	if _, err := os.Stat(pki_base); os.IsNotExist(err) {
		os.Mkdir(pki_base, 0700)
	}

	if _, err := os.Stat(path.Join(pki_base, key_file)); os.IsNotExist(err) {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			log.Fatalln(err)
		}
		prvKey = key
		key_bytes, key_err := x509.MarshalECPrivateKey(prvKey)
		if key_err != nil {
			log.Fatalln(key_err)
		}
		werr := os.WriteFile(path.Join(pki_base, key_file), key_bytes, 0600)
		if werr != nil {
			log.Fatalln(werr)
		}
	} else {
		pbytes, perr := os.ReadFile(path.Join(pki_base, key_file))
		if perr != nil {
			log.Fatalln(perr)
		}
		key, kerr := x509.ParseECPrivateKey(pbytes)
		if kerr != nil {
			log.Fatalln(kerr)
		}
		prvKey = key
	}

	if _, err := os.Stat(path.Join(pki_base, aes_file)); os.IsNotExist(err) {
		encKey = make([]byte, 16)
		if _, err := rand.Read(encKey); err != nil {
			log.Fatalln(err)
		}
		block := &pem.Block{
			Type:  "AES-128 KEY",
			Bytes: encKey,
		}
		file, err := os.Create(path.Join(pki_base, aes_file))
		if err != nil {
			log.Fatalln(err)
		}
		defer file.Close()
		err = pem.Encode(file, block)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		aesbytes, err := os.ReadFile(path.Join(pki_base, aes_file))
		if err != nil {
			log.Fatalln(err)
		}
		b, _ := pem.Decode(aesbytes)
		encKey = b.Bytes
	}
}

func Sign(bytes []byte) ([]byte, error) {
	result := make([]byte, 0)
	hash := sha256.Sum256(bytes)
	sig, sig_err := ecdsa.SignASN1(rand.Reader, prvKey, hash[:])
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
	hash := sha256.Sum256(payload)

	if ecdsa.VerifyASN1(&prvKey.PublicKey, hash[:], sig) {
		return payload, nil
	}
	return nil, errors.New("invalid signature")
}
