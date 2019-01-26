package oauth

import (
	"net/http"
)

type ResponseData map[string]interface{}
type ResponseType int

const (
	DATA ResponseType = iota
	REDIRECT
)

type Response struct {
	Type		   ResponseType
	StatusCode	   int
	StatusText 	   string
	ErrorStatusCode    int
	URL		   string
	Headers		   http.Header
	IsError 	   bool
	ErrorId 	   string
	InternalError      error
	RedirectInFragment bool
	Output 		   ResponseData
}

func NewResponse() *Response {
	r := &Response{
		Type:	         DATA,
		StatusCode:      200,
		ErrorStatusCode: 200,
		Headers:         make(http.Header),
		Output:          make(ResponseData),
		IsError:         false,
	}

	r.Headers.Add(
		"Cache-Control",
		"no-cache, no-store, max-age=0, must-revalidate",
	)
	r.Headers.Add(
		"Pragma",
		"no-cache",
	)
	r.Headers.Add(
		"Expires",
		"Fri, 01 Jan 1990 00:00:00 GMT",
	)

	return r
}

func (r *Response) Close() {

}

func (r *Response) SetErrorState(id string, description string, state string) {

}

func(r *Response) SetErrorUri(id string, description string, uri string, state string) {

	r.IsError = true
	r.ErrorId = id
	r.StatusCode = r.ErrorStatusCode

	if r.StatusCode != 200 {
		r.StatusText = description
	}

}