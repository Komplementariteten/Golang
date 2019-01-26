package token

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"strings"
	"ventose.cc/tools"
)

const (
	TOKEN_JWT       = "JWT"
	TOKEN_FVT       = "FVT"
	ALGORITHM_HS256 = "HS256"
	ALGORITHM_HS384 = "HS384"
	ALGORITHM_HS512 = "HS512"
)

type Token interface {
	GetHeader() *TokenHeader
	GetPayLoad() TPayLoadInterface
	String(key []byte) (string, error)
	Verify(token string, key []byte) (bool, error)
	SetPayLoad(p TPayLoadInterface) error
	LoadPayLoad(b64 string) error
	LoadHeader(header *TokenHeader)
}

type TokenHeader struct {
	Algorithm string `json:"alg"`
	TokenType string `json:"typ"`
	THeaderInterface
}

type THeaderInterface interface {
	ToJson() ([]byte, error)
	GetAlgorithm() hash.Hash
}

type TPayLoadInterface interface {
	ToJson() ([]byte, error)
}

func NewToken(typ string) Token {
	switch typ {
	case TOKEN_JWT:
		return newJWT()
	case TOKEN_FVT:
		return newFVT()
	default:
		panic("unkown type")
	}
}

func LoadToken(t string) (Token, error) {
	parts := strings.Split(t, ".")
	if len(parts) != 3 {
		fmt.Errorf("Token must contain 3 Parts got %d", len(parts))
	}
	orgHeader, err := tools.Base64Encode.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	header := &TokenHeader{}
	err = json.Unmarshal(orgHeader, header)
	if err != nil {
		return nil, err
	}

	nt := NewToken(header.TokenType)
	nt.LoadHeader(header)
	err = nt.LoadPayLoad(parts[1])
	if err != nil {
		return nil, err
	}
	return nt, nil
}

func VerifyToken(tokenstring string, key []byte) (bool, error) {
	t, err := LoadToken(tokenstring)
	if err != nil {
		return false, err
	}
	signVerify, err := t.Verify(tokenstring, key)
	if err != nil {
		return false, err
	}
	if !signVerify {
		return false, errors.New("Failed to Verify Signature")
	}
	return true, nil
}

func getAlgorithem(name string) hash.Hash {
	switch name {
	case ALGORITHM_HS256:
		return sha256.New()
	case ALGORITHM_HS384:
		return sha512.New384()
	case ALGORITHM_HS512:
		return sha512.New()
	default:
		return sha256.New()
	}
}

func getAlgorithemFromBase64Header(b64 string, k []byte) (hash.Hash, error) {
	hb, err := tools.Base64Encode.DecodeString(b64)
	if err != nil {
		return nil, err
	}
	header := &TokenHeader{}
	err = json.Unmarshal(hb, header)
	if err != nil {
		return nil, err
	}
	return hmac.New(header.GetAlgorithm, k), nil
}

func verify(t string, k []byte) (check bool, err error) {

	parts := strings.Split(t, ".")
	if len(parts) != 3 {
		fmt.Errorf("Token must contain 3 Parts got %d", len(parts))
	}
	alg, err := getAlgorithemFromBase64Header(parts[0], k)
	message := parts[0] + "." + parts[1]
	//alg.
	alg.Write([]byte(message))
	checkToken := alg.Sum(nil)
	orgToken, err := tools.Base64Encode.DecodeString(parts[2])

	if !bytes.Equal(checkToken, orgToken) {
		return false, fmt.Errorf("Token verfication failed")
	}

	return true, nil
}

func stringify(t Token, k []byte) (token string, err error) {
	b, err := sign(t, k)
	if err != nil {
		return
	}
	hj, err := t.GetHeader().ToJson()
	if err != nil {
		return
	}
	pj, err := t.GetPayLoad().ToJson()
	if err != nil {
		return
	}
	p1 := tools.Base64Encode.EncodeToString(hj)
	p2 := tools.Base64Encode.EncodeToString(pj)

	token = p1 + "." + p2 + "." + tools.Base64Encode.EncodeToString(b)
	return
}

func sign(t Token, k []byte) (token []byte, err error) {
	h := t.GetHeader()
	p := t.GetPayLoad()
	if p == nil {
		return nil, fmt.Errorf("No Payload in token %t -> %v", t, p)
	}
	hb, err := h.ToJson()
	if err != nil {
		return
	}
	p1 := tools.Base64Encode.EncodeToString(hb)
	pb, err := p.ToJson()
	if err != nil {
		return
	}
	p2 := tools.Base64Encode.EncodeToString(pb)

	message := p1 + "." + p2
	messageBytes := []byte(message)

	mac := hmac.New(h.GetAlgorithm, k)
	mac.Write(messageBytes)
	token = mac.Sum(nil)
	return
}

/*
TokenHeader
*/
func (h *TokenHeader) ToJson() (bytes []byte, err error) {
	bytes, err = json.Marshal(h)
	return
}

func (h *TokenHeader) GetAlgorithm() hash.Hash {
	return getAlgorithem(h.Algorithm)
}
