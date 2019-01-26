package oauth

import (
	"net/http"
	"net/url"
)

type RedirectEndpoint struct {
}

func (t *RedirectEndpoint) Handle(w http.ResponseWriter, r *http.Request) {
}
