package token

import (
	"encoding/json"
	"errors"
	"time"
	"ventose.cc/tools"
)

type JWT struct {
	Token
	Header    *TokenHeader
	Payload   TPayLoadInterface
	signature []byte
}

type JWTPayLoad struct {
	Issuer         string `json:"iss"`
	Subject        string `json:"sub"`
	Audience       string `json:"aud"`
	ExpirationTime int    `json:"exp"`
	NotBefore      int    `json:"nbf"`
	IssuedAt       int    `json:"iat"`
	JWTId          string `json:"jti"`
	TPayLoadInterface
}

/*
JWT
*/
func newJWT() *JWT {
	token := &JWT{}
	token.Header = &TokenHeader{}
	token.Header.Algorithm = ALGORITHM_HS256
	token.Header.TokenType = "JWT"
	return token
}

func (t *JWT) SetPayLoad(p TPayLoadInterface) error {

	jwtp, ok := p.(*JWTPayLoad)
	if ok {
		jwtp.IssuedAt = int(time.Now().Unix())
		jwtp.NotBefore = int(time.Now().Unix())
		t.Payload = jwtp
	} else {
		return errors.New("Payload Can't bee converted to JWTPayLoad")
	}
	return nil
}

func (t *JWT) GetHeader() *TokenHeader {
	if t.Header != nil {
		return t.Header
	}
	t.Header = &TokenHeader{}
	t.Header.Algorithm = ALGORITHM_HS256
	t.Header.TokenType = "JWT"
	return t.Header
}
func (t *JWT) GetPayLoad() TPayLoadInterface {
	if t.Payload != nil {
		return t.Payload
	}
	t.Payload = &JWTPayLoad{}
	return t.Payload
}
func (t *JWT) String(k []byte) (token string, err error) {
	return stringify(t, k)
}

func (t *JWT) Verify(tokenString string, key []byte) (check bool, err error) {
	ok, err := verify(tokenString, key)
	if err != nil || !ok {
		return false, err
	}
	p, ok := t.Payload.(*JWTPayLoad)
	if !ok {
		return false, errors.New("Wrong type of Payload")
	}
	c := int(time.Now().Unix())
	if c > p.ExpirationTime {
		return false, errors.New("Token has expired")
	}
	if c < p.NotBefore {
		return false, errors.New("Token is not valid, jet")
	}
	return true, nil
}

func (t *JWT) LoadPayLoad(b64 string) error {

	pb, err := tools.Base64Encode.DecodeString(b64)
	if err != nil {
		return err
	}
	p := &JWTPayLoad{}
	err = json.Unmarshal(pb, p)
	if err != nil {
		return err
	}
	t.Payload = p
	return nil
}

func (t *JWT) LoadHeader(header *TokenHeader) {
	t.Header = header
}

/*
JSWTPayload
*/
func (p *JWTPayLoad) ToJson() (bytes []byte, err error) {
	bytes, err = json.Marshal(p)
	return
}
