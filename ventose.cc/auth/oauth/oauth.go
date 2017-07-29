package oauth

import (
	"fmt"
	"net/http"
	"strings"
	"ventose.cc/tools"
	"crypto/sha256"
	"encoding/binary"
	"crypto/hmac"
	"strconv"
	"ventose.cc/auth/oauth/token"
	"errors"
	"time"
	"github.com/cupcake/rdb/nopdecoder"
)

const AUTH_PATH 	= "/auth/"
const TOKEN_PATH 	= "/token"
const REDIRECT_PATH 	= "/redirect"
const REQUESTIDSIZE	= 32
const SESSIONLIFETIME   = 5

type OAuth struct {
	AuthEndpoint 		*AuthorizationEndpoint
	TokenEndpoint 		*TokenEndpoint
	RedirectEndpoint	*RedirectEndpoint
	Clients map[string]*Client
	http.Handler
	Initialized bool
	Store *Store
	sessionName string
}

func GetHandler() OAuth {
	var h OAuth
	h.Initialized = false
	err := h.LoadConfiguration()
	if err != nil {
		panic(err)
	}
	return h
}

func (s *OAuth) LoadConfiguration() error {

	if !s.Initialized {
		s.AuthEndpoint = &AuthorizationEndpoint{}
		s.AuthEndpoint.LoginView = nil
		s.RedirectEndpoint = &RedirectEndpoint{}
		s.TokenEndpoint = &TokenEndpoint{}
		store, err := getStore()
		if err != nil {
			return err
		}
		s.Store = store
		clients, err := s.Store.GetClients()
		if err != nil {
			return err
		}
		s.Clients = clients
		s.sessionName = tools.GetRandomAsciiString(REQUESTIDSIZE)
		s.Initialized = true
	}

	return nil
}

func (s *OAuth) updateEndpointRequestFromFVT(er *Session, tokenString string) (*Session, error) {
	er.FormTokenText = tokenString
	t, err := token.LoadToken(tokenString)
	if err != nil {
		return nil, err
	}
	fvt, ok := t.(*token.FVT)
	if !ok {
		return nil, errors.New("Token is no FVT Token")
	}
	payload, ok := fvt.Payload.(*token.FVTPayLoad)
	if !ok {
		return nil, errors.New("Falied to load Payload Data")
	}
	if !s.HasSession(er) {
		return nil, errors.New("Session from POST Request not active on Server")
	}

	if er.ClientId != payload.ClientId {
		return nil, errors.New("Session Client and Payload Client don't match")
	}
	er.State = payload.State
	er.ClientId = payload.ClientId
	er.FormToken = fvt
	return er, nil
}

func (s *OAuth) getEndpointRequest(w http.ResponseWriter, r *http.Request) ( *Session, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	sc, err := r.Cookie(s.sessionName)
	if err != nil {
		return nil, err
	}
	er := &Session{}
	now := time.Now()
	// New Session
	if sc == nil {
		er.ClientId = r.Form.Get("client_id")
		er.State = r.Form.Get("state")
		er.RequestFailes = 0
		er.RequestId = tools.GetRandomAsciiString(REQUESTIDSIZE)
	} else if sc == nil &&  r.Method == http.MethodPost {
		er, err = s.Store.GetSession(sc.Value)
		if err != nil {
			return nil, err
		}
		if er == nil {
			return nil, errors.New("Can't POST without Session")
		}
	} else if sc != nil {
		if time.Now().Before(sc.Expires) {
			ec := &http.Cookie{Name: s.sessionName, Value: nil, Path: r.URL.Path, Expires: now.Add(-1 * time.Hour)}
			http.SetCookie(w, ec)
		} else {
			er, err = s.Store.GetSession(sc.Value)
		}
	}

	if r.Method == http.MethodGet && s.HasClient(er.ClientId) {
		c := &http.Cookie{Name: s.sessionName, Value: er.RequestId, Path: r.URL.Path, Expires: now.Add(SESSIONLIFETIME * time.Minute)}
		http.SetCookie(w, c)
		redirect_target := r.Form.Get("redirect_uri")
		if len(redirect_target) > 1 {
			er.RedirectUri = redirect_target
		}
	}
	er.Scope = r.Form.Get("scope")
	er.GrantType = r.Form.Get("grant_type")
	er.ResponseType = r.Form.Get("response_type")
	er.Code = r.Form.Get("code")
	ua := r.Header.Get("User-Agent")
	ra := r.RemoteAddr

	if ua == "" && ra == "" {
		er.UserAgent = ua
		er.RemoteAddr = ra
	} else if ua != er.UserAgent {
		return nil, errors.New("Session Ended, Useragend changed")
	} else {
		er.RemoteAddr = ra
	}

	s.Store.SetSession(er)

	return er, nil
}

func (s OAuth) ServeHTTP(w http.ResponseWriter, r *http.Request){
	er, err := s.getEndpointRequest(w, r)
	if err != nil {
		http.Error(w, "Failed to Parse Request with " + err.Error(), 420)
		return
	}
	client, err := s.ResolveClient(er.ClientId)
	if err != nil {
		http.Error(w, "Failed to get Clients", 420)
		return
	}
	if client == nil {
		http.NotFound(w,r)
	}

	if strings.Contains(r.RequestURI, AUTH_PATH){
		s.AuthEndpoint.Handle(w, r, er, client)
	} else if strings.Contains(r.RequestURI, TOKEN_PATH){
		s.TokenEndpoint.Handle(w, r)
	} else if strings.Contains(r.RequestURI, REDIRECT_PATH) {
		s.RedirectEndpoint.Handle(w, r)
	}
}

func (s *OAuth) HandleAuthorization(w http.ResponseWriter, r *http.Request) {

}