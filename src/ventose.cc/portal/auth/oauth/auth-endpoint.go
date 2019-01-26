package oauth

import (
	"fmt"
	"golang.org/x/tools/go/gcimporter15/testdata"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"ventose.cc/auth/oauth/token"
	"ventose.cc/tools"
)

const (
	AUTHORIZATION_CODE = "authorization_code"
	REFRESH_TOKEN      = "refresh_token"
	PASSWORD           = "password"
	CLIENT_CREDENTIALS = "client_credentials"
	ASSERTION          = "assertion"
	IMPLICIT           = "__implicit"
	LOGIN_PATH         = "login"
)

type TemplateData struct {
	LoginTarget        string
	FormToken          string
	LoginFieldName     string
	LoginFieldTitle    string
	PasswordFieldName  string
	PasswordFieldTitle string
	SubmitName         string
}

type AuthorizationEndpoint struct {
	LoginView       *template.Template
	BaseEndpointUrl url.URL
}

func preFillTemplateData() *TemplateData {
	Data := &TemplateData{}
	Data.LoginFieldName = tools.GetRandomAsciiString(12)
	Data.LoginFieldTitle = "Login:"
	Data.PasswordFieldName = tools.GetRandomAsciiString(12)
	Data.PasswordFieldTitle = "Password:"
	Data.SubmitName = "Login"
	return Data
}

func (s *AuthorizationEndpoint) SetLoginTemplate(templateString string) error {
	t, err := template.New("LoginView").Parse(templateString)
	if err != nil {
		panic(err)
	}
	s.LoginView = t
	return nil
}

func (s *AuthorizationEndpoint) SetLoginTemplateFile(templateFile string) error {

	m, err := os.Getwd()

	templateFile = m + "/" + templateFile

	if _, e := os.Stat(templateFile); os.IsNotExist(e) {
		//s.l.Panicf("Login Template File %v does not exist", templateFile)
		return fmt.Errorf("Login Template File %v does not exist: %v", templateFile, e)
	}

	t, err := template.ParseFiles(templateFile)
	if err != nil {
		panic(err)
	}
	s.LoginView = t
	return nil
}

func (ae *AuthorizationEndpoint) handleLoginView(w http.ResponseWriter, s *Session, client *Client) {
	t, err := client.GetFormToken(s)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	td := preFillTemplateData()
	td.FormToken = t
	td.LoginTarget = AUTH_PATH + LOGIN_PATH

	if ae.LoginView != nil {
		ae.LoginView.Execute(w, td)
	} else {
		fmt.Fprintf(w, "<h1>Hallo Welt</h1>")
	}
}
func (ae *AuthorizationEndpoint) handleLogin(w http.ResponseWriter, r *http.Request, s *Session, client *Client) {
	//csrfChallenge :=

	if len(s.FormTokenText) < 128 {
		http.Error(w, "CSRFToken not found", 403)
		return
	}

	check, err := token.VerifyToken(s.FormTokenText, client.Key)
	if err != nil {
		http.Error(w, "Token Verifycation faulted", 500)
	}

	time.Sleep(time.Duration(s.RequestFailes+1) * time.Second)

	if !check {
		s.RequestFailes++
		e := client.AddToUrl("msg", "You are not allowed to Request")
		if e != nil {
			http.Redirect(w, r, client.ClientUri.RequestURI(), http.StatusFound)
			return
		} else {
			http.Error(w, "Failed to add Message, Server Error", 500)
		}
	}

	p, ok := s.FormToken.Payload.(*token.FVTPayLoad)
	if !ok {
		http.Error(w, "CSRFToken not found", 403)
		return
	}
	//challenge := p.CSRFToken

}

func (ae *AuthorizationEndpoint) Handle(w http.ResponseWriter, r *http.Request, s *Session, client *Client) {

	if r.Method == http.MethodGet && strings.Contains(r.RequestURI, AUTH_PATH+LOGIN_PATH) {
		ae.handleLoginView(w, s, client)
	} else if r.Method == http.MethodPost && strings.Contains(r.RequestURI, AUTH_PATH+LOGIN_PATH) {
		ae.handleLogin(w, r, s, client)
	}

}
