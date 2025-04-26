package session

import (
	"FileBrowser/common"
	"FileBrowser/pki"
	"bytes"
	"compress/lzw"
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strings"
)

func SessionHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.ToLower(r.URL.Path)
	clean := strings.Replace(path, common.SESSION_PATH, "", 1)

	if clean == "new" {
		s := newSession()
		base := strings.Replace(path, "new", "", 1)
		session_sig := ToUrlPath(s)
		target := base + session_sig
		cookie := http.Cookie{
			Name:     common.SessionCookieName,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			Value:    session_sig,
		}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, target, http.StatusFound)
	} else {
		session_sig := strings.Replace(r.URL.Path, common.SESSION_PATH, "", 1)
		s, serr := common.SessionFromUrl(r.URL)
		if serr != nil {
			log.Fatalln(serr)
		}

		cookie := http.Cookie{
			Name:     common.SessionCookieName,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			Value:    session_sig,
		}
		http.SetCookie(w, &cookie)

		target := common.LIST_PATH + s.ContentListId
		http.Redirect(w, r, target, http.StatusSeeOther)
	}
}

func newSession() *common.Session {
	sessionId := uuid.New().String()
	session := &common.Session{
		Id: sessionId,
	}
	c := common.GetContentList(session)
	session.ContentListId = c.Id
	return session
}

func ToUrlPath(s *common.Session) string {
	var buf bytes.Buffer
	zw := lzw.NewWriter(&buf, lzw.MSB, 8)
	bytes, merr := json.Marshal(s)
	if merr != nil {
		log.Fatal(merr)
	}
	sig, sigerr := pki.Sign(bytes)
	if sigerr != nil {
		log.Fatal(sigerr)
	}
	_, gzer := zw.Write(sig)
	if gzer != nil {
		log.Fatal(gzer)
	}

	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}
	bto_encode := buf.Bytes()
	return base64.RawURLEncoding.EncodeToString(bto_encode)
}
