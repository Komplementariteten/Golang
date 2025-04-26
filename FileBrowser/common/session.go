package common

import (
	"FileBrowser/pki"
	"bytes"
	"compress/lzw"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

const SessionCookieName = "_owuid"

type Session struct {
	Id            string   `json:"id"`
	ContentListId string   `json:"content_list_id"`
	CurrentPath   []string `json:"current_path"`
}

func SessionFromCookie(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return nil, err
	}
	cerr := cookie.Valid()
	if cerr != nil {
		return nil, cerr
	}

	return loadSession(cookie.Value)
}

func loadSession(sig string) (*Session, error) {
	b, b_err := base64.RawURLEncoding.DecodeString(sig)
	if b_err != nil {
		return nil, b_err
	}
	zr := lzw.NewReader(bytes.NewBuffer(b), lzw.MSB, 8)
	var buf bytes.Buffer
	_, rerr := buf.ReadFrom(zr)
	if rerr != nil {
		return nil, rerr
	}
	sbytes := buf.Bytes()
	json_bytes, verr := pki.Verify(sbytes)
	if verr != nil {
		return nil, verr
	}
	session := Session{}
	jerr := json.Unmarshal(json_bytes, &session)
	if jerr != nil {
		return nil, jerr
	}
	return &session, nil
}

func SessionFromUrl(url *url.URL) (*Session, error) {
	b64 := strings.Replace(url.Path, SESSION_PATH, "", 1)
	return loadSession(b64)
}
