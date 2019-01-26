package token

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"ventose.cc/tools"
)

const TOKENLIVETIME = 960

type FVT struct {
	Token
	Header  *TokenHeader
	Payload TPayLoadInterface
}

type FVTPayLoad struct {
	ClientId       string `json:"rqi"`
	State          string `json:"stt"`
	ExpirationTime int    `json:"exp"`
	NotBefore      int    `json:"nbf"`
	IssuedAt       int    `json:"iat"`
	CSRFToken      string `json:"cft"`
	TPayLoadInterface
}

func newFVT() *FVT {
	token := &FVT{}
	token.Header = &TokenHeader{}
	token.Header.Algorithm = ALGORITHM_HS256
	token.Header.TokenType = "FVT"
	return token
}

func NewFvtTokenData(req string, state string, CSRFToken string, issuedAt int) *FVTPayLoad {
	p := &FVTPayLoad{}
	p.ClientId = req
	p.State = state
	p.CSRFToken = CSRFToken
	p.IssuedAt = issuedAt
	return p
}

/*
FVT
*/

func (t *FVT) SetPayLoad(p TPayLoadInterface) error {

	fvtp, ok := p.(*FVTPayLoad)
	if ok {
		if len(fvtp.CSRFToken) < 5 {
			return errors.New("CSRFToken not set")
		}
		if len(fvtp.ClientId) < 2 {
			return errors.New("No Client set!")
		}
		if len(fvtp.State) < 3 {
			return errors.New("Wrong or no State set")
		}
		fvtp.NotBefore = int(time.Now().Unix())
		fvtp.ExpirationTime = int(time.Now().Unix()) + TOKENLIVETIME
		t.Payload = fvtp
	} else {
		return fmt.Errorf("Can't set correct Payload for FVToken %v", p)
	}
	return nil
}

func (t *FVT) GetHeader() *TokenHeader {
	if t.Header != nil {
		return t.Header
	}
	t.Header = &TokenHeader{}
	t.Header.Algorithm = ALGORITHM_HS256
	t.Header.TokenType = "JWT"
	return t.Header
}
func (t *FVT) GetPayLoad() TPayLoadInterface {
	if t.Payload != nil {
		return t.Payload
	}
	t.Payload = &JWTPayLoad{}
	return t.Payload
}
func (t *FVT) String(k []byte) (token string, err error) {
	return stringify(t, k)
}

func (t *FVT) Verify(tokenString string, key []byte) (check bool, err error) {
	ok, err := verify(tokenString, key)
	if err != nil || !ok {
		return false, err
	}
	p, ok := t.Payload.(*FVTPayLoad)
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

func (t *FVT) LoadPayLoad(b64 string) error {
	pb, err := tools.Base64Encode.DecodeString(b64)
	if err != nil {
		return err
	}
	p := &FVTPayLoad{}
	err = json.Unmarshal(pb, p)
	if err != nil {
		return err
	}
	t.Payload = p
	return nil

	return nil
}

func (t *FVT) LoadHeader(header *TokenHeader) {
	t.Header = header
}

/*
JSWTPayload
*/
func (p *FVTPayLoad) ToJson() (bytes []byte, err error) {
	bytes, err = json.Marshal(p)
	return
}
