package auth

import (
	"crypto/rsa"
	"time"
	"ventose.cc/data"
)

type User struct {
	data.RequestContent
	Login    string `search:"userlogin"`
	Name     string `search:"username"`
	PassHash []byte
	Created  time.Time
	Id       []byte
	Key      *rsa.PrivateKey
}
