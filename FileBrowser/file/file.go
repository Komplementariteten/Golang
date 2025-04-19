package file

import (
	"FileBrowser/common"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func FileHandle(w http.ResponseWriter, r *http.Request) {
	s, serr := common.SessionFromCookie(r)
	if serr != nil {

	}
	c := common.GetContentList(s)
	if r.Method == "GET" {
		f, derr := readFile(r.URL.Path)
		if derr != nil {
			log.Fatalln(derr)
			return
		}
		
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(r.URL.Path)+"\"")
		io.Copy(w, f)
	}

	if r.Method == "POST" {
		ferr := r.ParseMultipartForm(600 << 20)
		if ferr != nil {
		}
		file, handler, uerr := r.FormFile("file")
		if uerr != nil {
		}
		defer file.Close()

	}
}

func storeFile() {

}

func storeFile(desc *multipart.FileHeader, list string) {

}

func readFile(path string) (r io.Reader, err error) {

}
