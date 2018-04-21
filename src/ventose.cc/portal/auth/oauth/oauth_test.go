package oauth

import (
	"testing"
	"net/http"
	"io/ioutil"
	"math/rand"
	"strings"
	"compress/gzip"
	"encoding/xml"
	"net/url"
	"ventose.cc/auth/oauth/token"
)
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const noTemplateString = "<h1>Hallo Welt</h1>"
const LoginTemplateStirng = `<form method="POST" action="{{.LoginTarget}}" >
<input type="text" name="{{.LoginFieldName}}" id="oalfventose" >D</input>
<input type="password" name="{{.PasswordFieldName}}" id="oapfventose" >C</input>
<input type="hidden" name="fvt" value="{{.FormToken}}">B</input>
<input type="submit" name="submit" value="{{.SubmitName}}">A</input>
</form>
`

type Html struct {
	XMLName xml.Name `xml:"form"`
	Method string   `xml:"method,attr"`
	Action string   `xml:"action,attr"`
	Items []children `xml:"input"`
}
type children struct {
	XMLName xml.Name `xml:"input"`
	Type string `xml:"type,attr"`
	Name string `xml:"name,attr"`
	Value string `xml:"value,attr"`
	Id string `xml:"id,attr"`
}

func startHttpServer(){
	go func(){
		http.ListenAndServe(":33445", nil)
	}()
}
func RandStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63() % int64(len(letterBytes))]
	}
	return string(b)
}

func GetClient(t *testing.T) *Client {

	c, err := NewClient("testclient", "http://www.ventose.cc")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func GetStringFromHttp(urlString string, t *testing.T) string {
	resp, err := http.Get(urlString)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		t.Error("Failed ", resp.StatusCode, resp.Status)
	}

	encHeader := resp.Header.Get("Content-Encoding")

	var loginformhtml string

	if strings.Contains(encHeader, "gzip") {
		gzipr, err := gzip.NewReader(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer gzipr.Close()
		rbytes, err := ioutil.ReadAll(gzipr)
		if err != nil {
			t.Fatal(err)
		}
		loginformhtml = string(rbytes[:])

	} else {
		rbytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		loginformhtml = string(rbytes[:])
	}
	return loginformhtml
}

func TestOAuth_NewClient(t *testing.T) {
	c := GetClient(t)
	_, err := c.ValidateURL(c.ClientUri.String())
	if err != nil {
		t.Fatal(err)
	}
}

func TestOAuth_AddClient(t *testing.T) {
	h := GetHandler()
	c := GetClient(t)
	h.AddClient(c)
}

func TestAuthorizationEndpoint_LoginForm(t *testing.T) {

	randString := RandStringBytesRmndr(8)

	h := GetHandler()

	c := GetClient(t)
	h.AddClient(c)
	http.Handle("/"+randString+"/", h)
	startHttpServer()
	loginformhtml := GetStringFromHttp("http://localhost:33445/"+randString+"/auth/dgdgertgdrt?client_id=" + c.ClientId + "&state=fdfd&redirect_url="+c.ClientUri.String()+"&scope=134435", t)

	if len(loginformhtml) == 0 {
		t.Error("Got no proper Response Data")
	}

	if loginformhtml != noTemplateString {
		t.Error("Auth Endpoint return " + loginformhtml + " but noTemplateString is " + noTemplateString)
	}

	err := h.AuthEndpoint.SetLoginTemplate(LoginTemplateStirng)
	if err != nil {
		t.Error(err)
	}

	loginurl := "http://localhost:33445/"+randString+"/auth/dgdgertgdrt?client_id=" + c.ClientId + "&state=fdfd&redirect_url="+c.ClientUri.String()+"&scope=134435"

	loginformhtml = GetStringFromHttp(loginurl, t)

	if len(loginformhtml) == 0 {
		t.Error("Got no proper Response Data")
	}

	if loginformhtml == noTemplateString {
		t.Error("Auth Endpoint return " + loginformhtml + " but noTemplateString is " + noTemplateString)
	}

	f := Html{Method:"none", Action:"none"}
	err = xml.Unmarshal([]byte(loginformhtml), &f)
	if err != nil {
		t.Fatal(err)
	}

	if f.Method != "POST" {
		t.Errorf("Loginform is not POST: %v", f)
	}

	_, err = url.Parse(f.Action)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(loginurl, f.Action) {
		t.Fatal("Form leads to the wrong action URL")
	}

	for _, input := range f.Items {
		if input.Name == "fvt" {
			ok, err := token.VerifyToken(input.Value, c.Key)
			if err != nil {
				t.Error(err)
			}
			if !ok {
				t.Error("Failed to Verify Token")
			}
		}
	}

}

func TestOAuth_ServeHTTP(t *testing.T) {

	randString := RandStringBytesRmndr(8)

	h := GetHandler()

	c := GetClient(t)
	h.AddClient(c)
	http.Handle("/"+randString+"/", h)
	startHttpServer()

	resp, err := http.Get("http://localhost:33445/"+randString+"/auth/dgdgertgdrt?client_id=" + c.ClientId + "&state=fdfd&redirect_url="+c.ClientUri.String()+"&scope=134435")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode > 399 {
		t.Error("Failed ", resp.StatusCode, resp.Status)
	}

}
