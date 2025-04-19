package common

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type FileType int64

const (
	FileTypeFile FileType = iota
	FileTypeDirectory
	FileTypeUnkown
)

var lock = &sync.Mutex{}

var contentLists []*ContentList

func readOrNewContentList(session *Session) *ContentList {
	if session.ContentListId == "" {
		session.ContentListId = uuid.New().String()
	}
	c, new_err := readContentList(session.ContentListId)
	if new_err != nil {
		log.Println(new_err)
		c, new_err = newContentList(session.ContentListId)
		if new_err != nil {
			log.Fatal(new_err)
		}
	}
	return &c
}

func GetContentList(session *Session) *ContentList {
	// Initialize
	if contentLists == nil {
		c := readOrNewContentList(session)
		lock.Lock()
		contentLists = append(contentLists, c)
		lock.Unlock()
		return c
	}
	lock.Lock()
	defer lock.Unlock()

	// Look for a loaded one
	for _, c := range contentLists {
		if c.Id == session.ContentListId {
			return c
		}
	}
	c := readOrNewContentList(session)
	contentLists = append(contentLists, c)
	return c
}

func newContentList(id string) (ContentList, error) {
	if _, err := os.Stat(id); os.IsNotExist(err) {
		os.Mkdir(id, 0700)
	}

	file, fp_err := CreateOrOpenFile(filepath.Join("."+LIST_PATH, id))
	if fp_err != nil {
		return ContentList{}, fp_err
	}
	defer file.Close()

	newList := ContentList{
		Entries:    make([]ListEntry, 0),
		StorageDir: id,
		Id:         id,
	}
	b, j_err := json.Marshal(newList)
	if j_err != nil {
		return ContentList{}, j_err
	}
	_, w_err := io.Copy(file, bytes.NewReader(b))
	if w_err != nil {
		return ContentList{}, w_err
	}

	return newList, nil
}

func readContentList(path string) (ContentList, error) {
	if _, err := os.Stat(LIST_PATH); os.IsNotExist(err) {
		os.Mkdir(LIST_PATH, 0700)
	}

	file, fp_err := CreateOrOpenFile(filepath.Join(LIST_PATH, path))
	if fp_err != nil {

	}
	defer file.Close()

	bytes, read_err := io.ReadAll(file)
	if read_err != nil {
		return ContentList{}, read_err
	}
	var contentList ContentList
	j_err := json.Unmarshal(bytes, &contentList)
	if j_err != nil {
		return ContentList{}, j_err
	}
	return contentList, nil
}

type ContentList struct {
	Entries    []ListEntry `json:"entries"`
	StorageDir string      `json:"base_dir"`
	Id         string      `json:"id"`
}

type ListEntry struct {
	Path     string   `json:"path"`
	FileName string   `json:"file_name"`
	Size     int64    `json:"size"`
	Sig      string   `json:"sig"`
	Type     FileType `json:"type"`
	Id       string   `json:"id"`
	ParentId string   `json:"parent_id"`
}
