package oauth

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"ventose.cc/auth/oauth/token"
	"ventose.cc/tools"
)

type Client struct {
	ClientId     string
	Title        string
	Description  string
	ClientUri    *url.URL
	LoginName    string
	PasswordName string
	Key          []byte
}

func NewClient(name string, urlString string) (*Client, error) {
	c := &Client{}
	c.Title = name
	c.ClientId = tools.GetRandomAsciiString(10)
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		return nil, fmt.Errorf("URL has to be a absolute URL")
	}
	c.LoginName = tools.GetRandomAsciiString(20)
	c.PasswordName = tools.GetRandomAsciiString(20)
	c.ClientUri = u
	c.Key = tools.GetRandomBytes(256)
	return c, nil
}

func (c *Client) AddToUrl(key string, value string) error {
	var urlString = c.ClientUri.RequestURI()
	if strings.Contains(urlString, "?") {
		urlString += "&" + key + "=" + value
	} else {
		urlString += "?" + key + "=" + value
	}
	u, err := url.Parse(urlString)
	if err != nil {
		return err
	}
	c.ClientUri = u
	return nil
}

func (c *Client) ValidateURL(reqUrl string) (bool, error) {

	v, e := url.Parse(reqUrl)
	if e != nil {
		return false, e
	}
	if v.Host != c.ClientUri.Host || v.Scheme != c.ClientUri.Scheme {
		return false, fmt.Errorf("Host of target client URL don't match")
	}
	return true, nil
}

func (c *Client) GetFormToken(r *Session) (b64 string, err error) {

	t := token.NewToken(token.TOKEN_FVT)
	issuedAt := int(time.Now().Unix())
	stateKey := []byte(r.State)
	p := token.NewFvtTokenData(r.ClientId, r.State, r.GetCSRFRoken(stateKey, issuedAt), issuedAt)
	err = t.SetPayLoad(p)
	if err != nil {
		return
	}
	b64, err = t.String(c.Key)
	return
}

/* OAtuh */
func (s *OAuth) AddClient(client *Client) {

	if !s.HasClient(client.ClientId) {
		s.Clients[client.ClientId] = client
	}
}

func (s *OAuth) HasClient(client_id string) bool {
	_, ok := s.Clients[client_id]
	return ok
}

func (s *OAuth) ResolveClient(client_id string) (*Client, error) {
	if s.HasClient(client_id) {
		return s.Clients[client_id], nil
	}
	return nil, fmt.Errorf("Client not registered on Server")
}
