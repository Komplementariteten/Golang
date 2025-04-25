package file

import (
	"FileBrowser/common"
	"crypto/md5"
	"encoding/hex"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

func FileHandle(w http.ResponseWriter, r *http.Request) {
	s, serr := common.SessionFromCookie(r)
	if serr != nil {
		http.Error(w, serr.Error(), http.StatusForbidden)
	}
	c := common.GetContentList(s)
	if r.Method == "GET" {
		/* f, derr := readFile(r.URL.Path)
		if derr != nil {
			log.Fatalln(derr)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(r.URL.Path)+"\"")
		io.Copy(w, f) */
	}

	if r.Method == "POST" {
		ferr := r.ParseMultipartForm(600 << 20)
		if ferr != nil {
			http.Error(w, ferr.Error(), http.StatusBadRequest)
		}
		file, handler, uerr := r.FormFile("file")
		if uerr != nil {
			http.Error(w, uerr.Error(), http.StatusBadRequest)
		}
		defer file.Close()
		entry, serr := storeFile(file, handler)
		if serr != nil {
			http.Error(w, serr.Error(), http.StatusInternalServerError)
		}
		c.Entries = append(c.Entries, entry)

		merr := common.StoreContentList(c)
		if merr != nil {
			http.Error(w, merr.Error(), http.StatusInternalServerError)
		}
	}
}

func storeFile(file multipart.File, desc *multipart.FileHeader) (common.ListEntry, error) {
	println("Storing: " + desc.Filename)
	year := strconv.Itoa(time.Now().Year())
	month := time.Now().Month().String()
	folder := path.Join(year, month)

	err := common.CreateDirIfNotExists(folder)
	if err != nil {
		return common.ListEntry{}, err
	}
	filepath := path.Join(folder, desc.Filename)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		newfile, err := os.Create(filepath)
		if err != nil {
			return common.ListEntry{}, err
		}
		defer newfile.Close()
		if _, nerr := newfile.ReadFrom(file); nerr != nil {
			return common.ListEntry{}, nerr
		}
	}
	bytes, rerr := io.ReadAll(file)
	if rerr != nil {
		return common.ListEntry{}, rerr
	}
	check_sum := md5.Sum(bytes)
	return common.ListEntry{
		Path:     filepath,
		FileName: desc.Filename,
		Size:     desc.Size,
		Type:     common.FileTypeFile,
		Id:       uuid.New().String(),
		Sig:      hex.EncodeToString(check_sum[:]),
	}, nil
}

type FileAccess struct {
	Path      string
	SessionId uuid.UUID
	Signature []byte
}
