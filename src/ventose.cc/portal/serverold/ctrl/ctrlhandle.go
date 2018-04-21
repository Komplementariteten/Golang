package ctrl

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func SetupHeader(w http.ResponseWriter) {
	w.Header().Add("Content-type", "application/json; charset=utf-8")
}

func loadParameter(w http.ResponseWriter, parameterString string, v interface{}) error {
	cfgBytes, err := base64.StdEncoding.DecodeString(parameterString)
	if err != nil {
		return logAndError(w, "AppHandle - Can't Decode Paramert: %s", parameterString)
	}
	err = json.Unmarshal(cfgBytes, v)
	if err != nil {
		return logAndError(w, "AppHandle - Can't Regain from Paramert: %s", cfgBytes)
	}
	return nil
}

func logAndError(w http.ResponseWriter, format string, v ...interface{}) error {
	error := fmt.Errorf(format, v)
	logger.Println(error)
	http.Error(w, error.Error(), http.StatusInternalServerError)
	return error
}

func validateRequest(h func(r *Request, w http.ResponseWriter)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(w, r)
		valid := false

		if r.Method == "PUT" {
			valid = true
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Not Acceptable", err), http.StatusNotAcceptable)
			return
		}

		contentType := r.Header.Get("Content-type")

		//contentType = http.DetectContentType(body)

		if !mimeRegex.MatchString(contentType) {
			http.Error(w, contentType+" Not Acceptable", http.StatusNotAcceptable)
			return
		}
		request := new(Request)
		err = json.Unmarshal(body, request)
		if err != nil {
			http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
			return
		}

		hashHeader := r.Header.Get("X-Content-Hash")

		reqHash, _ := base64.StdEncoding.DecodeString(hashHeader)

		//hash := sha256.Sum256(body)
		hash := sha256.New()
		hashBytes := hash.Sum(body)

		if !bytes.Equal(hashBytes, reqHash) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		plainText, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, this.Key, request.Signature, []byte(""))

		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		var cmd, ts string
		fmt.Sscanf(string(plainText), "%s %s", &cmd, &ts)

		if cmd != request.Command {
			http.Error(w, "Forbidden", http.StatusForbidden)
			logger.Printf("Command in JSON is invalid, expected %s got %s in %s", cmd, request.Command, string(plainText))
			return
		}

		if request.Date.Format(time.RFC3339) != ts {
			http.Error(w, "Forbidden", http.StatusForbidden)
			logger.Printf("Time in JSON is invalid, expected %s got %s", ts, request.Date.String())
			return
		}

		if valid {
			h(request, w)
		}
	}
}

func logRequest(w http.ResponseWriter, r *http.Request) {

	logger.Printf("%s %s %d %s\n", r.Method, r.RequestURI, r.ContentLength, r.RemoteAddr)
}

type CtrlHandle struct {
}

func (h CtrlHandle) Default(r *Request, w http.ResponseWriter) {
	http.Error(w, "Not Fount", http.StatusNotFound)
}
