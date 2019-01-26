package oauth

import (
	"crypto/rand"
	"crypto/rsa"
	"html/template"
	"net/http"
	"ventose.cc/auth"
	"ventose.cc/data"
)

type Token struct {
	access_token  string
	token_type    string
	expires_in    int
	refresh_token string
}

type LoginTemplate struct {
	Prefix              string
	LoginUrl            string
	LoginTitle          string
	PasswordTitle       string
	ValidationTokenName string
	ValidationToken     string
	ChallengeResponse   string
	SubmitBtnName       string
	ResetBtnName        string
}

const loginForm = `<div class="{{.Prefix}}_loginForm">
<form method="POST" action="{{.LoginUrl}}" >
<div class="{{.Prefix}}_formrow" >
	<label >{{.LoginTitle}}<input type="text" name="Login" /></label>
</div>
<div class="{{.Prefix}}_formrow" >
	<label >{{.PasswordTitle}}<input type="password" name="Password" /></label>
</div>
{{if .Captcha}}
<div class="{{.Prefix}}_formrow" >
</div>
{{end}}
<div class="{{.Prefix}}_formrow" >
	<input type="hidden" name="challenge" value="{{.ChallengeResponse}}" />
	<input type="hidden" name="{{.ValidationTokenName}}" value="{{.ValidationToken}}" />
	<input type="submit" value="{{.SubmitBtnName}}" />
	<input type="reset" value="{{.ResetBtnName}}" />
</div>
</form>
</div>`

const tokenName = "__validationtoken"

var authFrontend *auth.Consumer
var templateData *LoginTemplate
var t *template.Template

func init() {
	authFrontend = new(auth.Consumer)
	authFrontend.Description = "OAuth"
	authFrontend.Key, _ = rsa.GenerateKey(rand.Reader, 2048)
	templateData = new(LoginTemplate)
	templateData.LoginTitle = "Login"
	templateData.PasswordTitle = "Password"
	templateData.ResetBtnName = "Reset"
	templateData.SubmitBtnName = "Login"
	templateData.ValidationTokenName = tokenName
	t := template.New("Login Template")
	t, _ = t.Parse(loginForm)
}

func Authorize(path string, ab *auth.AuthBackend, s *data.StorageConnection) http.HandlerFunc {
	ab.AddFrontend(authFrontend)
	templateData.Prefix = path
	templateData.LoginUrl = path
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			t.Execute(w, templateData)
		}
	}
}
