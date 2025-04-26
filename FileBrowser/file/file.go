package file

import (
	"FileBrowser/common"
	"FileBrowser/pki"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const FILE_DESC_LENGTH = 72

const UUID_LEN = 36

func FileHandle(w http.ResponseWriter, r *http.Request) {
	s, serr := common.SessionFromCookie(r)
	if serr != nil {
		http.Error(w, serr.Error(), http.StatusForbidden)
	}
	c := common.GetContentList(s)
	if r.Method == "GET" {
		file_id := strings.Replace(r.URL.Path, common.FILE_PATH, "", 1)
		cipher, derr := base64.RawURLEncoding.DecodeString(file_id)
		if derr != nil {
			http.Error(w, derr.Error(), http.StatusBadRequest)
		}
		plain, cerr := pki.Decrypt(cipher)
		if cerr != nil {
			http.Error(w, cerr.Error(), http.StatusBadRequest)
		}

		if len(plain) != FILE_DESC_LENGTH {
			http.Error(w, "", http.StatusForbidden)
		}

		file_id = string(plain[:UUID_LEN])
		list_id := string(plain[UUID_LEN:])

		if list_id != s.ContentListId {
			http.Error(w, "", http.StatusForbidden)
		}
		file := c.FindFile(file_id)
		f, derr := os.Open(file.Path)
		if derr != nil {
			http.Error(w, derr.Error(), http.StatusInternalServerError)
		}
		defer f.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+file.FileName+"\"")
		_, coerr := io.Copy(w, f)
		if coerr != nil {
			http.Error(w, coerr.Error(), http.StatusInternalServerError)
		}
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

		link, lerr := createLink(entry.Id, s.ContentListId)
		if lerr != nil {
			http.Error(w, lerr.Error(), http.StatusInternalServerError)
		}
		entry.Link = link
		c.Entries = append(c.Entries, entry)

		merr := common.StoreContentList(c)
		if merr != nil {
			http.Error(w, merr.Error(), http.StatusInternalServerError)
		}
	}
}

func createLink(file_id string, collection_id string) (string, error) {
	id_bytes := []byte(file_id)
	list_bytes := []byte(collection_id)
	plain_text := append(id_bytes, list_bytes...)
	cripher, err := pki.Encrypt(plain_text)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(cripher), nil
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
