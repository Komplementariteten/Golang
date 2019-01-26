package token

import (
	"testing"
	"ventose.cc/tools"
)

func TestJWT_SignBase64(t *testing.T) {
	key := tools.GetRandomBytes(12)
	token := NewToken(TOKEN_JWT)
	payload := &JWTPayLoad{}
	payload.Issuer = "me"
	payload.Audience = tools.GetRandomAsciiString(12)
	payload.JWTId = tools.GetRandomAsciiString(12)
	payload.Subject = tools.GetRandomAsciiString(12)
	token.SetPayLoad(payload)

	b64, err := token.String(key)
	if err != nil {
		t.Error(err)
	}

	check, err := token.Verify(b64, key)
	if err != nil {
		t.Error(err)
	}
	if !check {
		t.Error("Token Verification failed")
	}
}
