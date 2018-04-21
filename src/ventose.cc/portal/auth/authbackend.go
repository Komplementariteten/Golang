package auth

import (
	"bytes"
	"crypto/rsa"
	"ventose.cc/data"
)

const (
	AUTH_USER = iota
	AUTH_ADMIN
	AUTH_SU
)

type AuthBackend struct {
	storage        *data.StorageConnection
	StorageOnline  bool
	FrontendOnline bool
	Frontends      []Consumer
	AuthChannel    chan AuthRequest
	AddChannel     chan User
	DelChannel     chan User
	UpdateChannel  chan User
	Closer         chan bool
	Response       chan AuthResponse
	grandChannel   chan GrandRequest
	grandResponse  chan GrandResponse
}

type ConsumerInterface interface {
	ConnectAuthBackend(b *AuthBackend)
}

type Consumer struct {
	ConsumerInterface
	Id             []byte
	Name           string
	Description    string
	Key            *rsa.PrivateKey
	GrandRequester chan GrandRequest
	GrandResponder chan GrandResponse
}

type ConsumerAccess struct {
	User        []byte
	Consumer    []byte
	Id          []byte
	Signatur    []byte
	AccessLevel int
}

type AuthResponse struct {
	Error       error
	Authorized  bool
	UserId      []byte
	AccessLevel int
}

type AuthRequest struct {
	Id         []byte
	Login      string
	Name       string
	Hash       []byte
	CunsumerId []byte
}

type GrandRequest struct {
	UserId     []byte
	CunsumerId []byte
}
type GrandResponse struct {
	UserId      []byte
	AccessLevel int
}

func NewAuthBackend() *AuthBackend {
	b := new(AuthBackend)
	b.StorageOnline = false
	b.AuthChannel = make(chan AuthRequest)
	b.AddChannel = make(chan User)
	b.DelChannel = make(chan User)
	b.UpdateChannel = make(chan User)
	b.Closer = make(chan bool)
	b.Response = make(chan AuthResponse)
	b.grandChannel = make(chan GrandRequest)
	return b
}

func (b *AuthBackend) AddFrontend(c *Consumer) {
	c.ConnectAuthBackend(b)
	go func(c *Consumer) {
		for gr := range c.GrandRequester {
			b.grandChannel <- gr
			rg := <-b.grandResponse
			c.GrandResponder <- rg
		}
	}(c)
}

func (b *AuthBackend) Run() {
	if b.FrontendOnline && len(b.Frontends) > 0 {
		go func() {
			b.serve()
		}()
	} else {
		panic("Auth Backend got no configured Frontend")
	}
}

func (b *AuthBackend) respond(r data.StorageResponse) {
	resp := new(AuthResponse)
	if r.Error != nil {
		resp.Error = r.Error
		resp.Authorized = false
	} else {
		if len(r.Affected) > 0 {
			resp.Authorized = true
			resp.UserId = r.Affected
		} else {
			resp.Authorized = false
		}
	}
	b.Response <- *resp
}

func (b *AuthBackend) serve() {
	for {
		select {
		case user := <-b.UpdateChannel:
			req := new(data.StorageRequest)
			req.Type = data.UpdateRequest
			req.Content = &user
			req.Element = user.Id
			b.storage.RequestChannel <- *req
			resp := <-b.storage.ResponseChannel
			b.respond(resp)
		case user := <-b.DelChannel:
			req := new(data.StorageRequest)
			req.Type = data.DeleteRequest
			req.Content = &user
			req.Element = user.Id
			b.storage.RequestChannel <- *req
			resp := <-b.storage.ResponseChannel
			b.respond(resp)
		case authreq := <-b.AuthChannel:
			req := new(data.StorageRequest)
			sr := new(User)
			if len(authreq.Id) > 0 {
				req.Type = data.ReadRequest
				req.Element = authreq.Id
				req.Content = sr
			} else {
				req.Type = data.SearchRequest
				sr.Name = authreq.Name
				sr.Login = authreq.Login
				req.Content = sr
			}
			b.storage.RequestChannel <- *req
			resp := <-b.storage.ResponseChannel
			if resp.Content == nil {
				b.respond(resp)
				break
			}

			user := resp.Content.(*User)
			if bytes.Equal(user.PassHash, authreq.Hash) {
				b.respond(resp)
			} else {
				resp.Affected = nil
				b.respond(resp)
			}
		case <-b.Closer:
			return
		case user := <-b.AddChannel:
			req := new(data.StorageRequest)
			req.Type = data.CreateRequest
			req.Content = &user
			b.storage.RequestChannel <- *req
			resp := <-b.storage.ResponseChannel
			b.respond(resp)
		case <-b.grandChannel:

		}
		/*defer func() {
			dresp := <-b.storage.ResponseChannel
			resp := new(AuthResponse)
			resp.Error = dresp.Error
			b.Response <- resp
		}()*/
	}
}

func (b *AuthBackend) ConnectStorage(s *data.StorageConnection) {
	b.storage = s
}
