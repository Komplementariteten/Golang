package oauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"strconv"
	"ventose.cc/auth/oauth/token"
)

type Session struct {
	ClientId      string
	State         string
	Scope         string
	UserAgent     string
	Cookies       string
	RequestId     string
	ResponseType  string
	Code          string
	GrantType     string
	RemoteAddr    string
	RedirectUri   string
	FormToken     *token.FVT
	FormTokenText string
	RequestFailes uint8
}

func (s *OAuth) HasSession(ses *Session) bool {
	if ses == nil {
		return false
	}
	return false
}

func (er *Session) GetCSRFRoken(key []byte, timestamp int) string {

	mac := hmac.New(sha256.New, key)

	strMessage := strconv.Itoa(timestamp) + "#" + er.UserAgent + "#" + er.RemoteAddr
	message := []byte(strMessage)
	mac.Write(message)
	csrfTokenBytes := mac.Sum(nil)

	csrfTokenInt := binary.BigEndian.Uint64(csrfTokenBytes)
	return strconv.FormatUint(csrfTokenInt, 16)
}
