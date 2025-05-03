package common

import (
	"FileBrowser/env"
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"path"
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
		c, new_err = newContentList(session.ContentListId)
		if new_err != nil {
			log.Fatal(new_err)
		}
	}
	return &c
}

func StoreContentList(contentList *ContentList) error {
	file_path := filepath.Join(env.Enviroment().BaseDir, LIST_PATH, contentList.Id)
	cbytes, cerr := json.Marshal(contentList)
	if cerr != nil {
		return cerr
	}

	werr := os.WriteFile(file_path, cbytes, 0600)
	if werr != nil {
		return werr
	}
	return nil
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
	listfile_path := path.Join(env.Enviroment().BaseDir, LIST_PATH)

	if _, err := os.Stat(listfile_path); os.IsNotExist(err) {
		os.Mkdir(id, 0700)
	}

	listFile := path.Join(listfile_path, id)

	newList := ContentList{
		Entries:    make([]ListEntry, 0),
		StorageDir: id,
		Id:         id,
	}
	b, j_err := json.Marshal(newList)
	if j_err != nil {
		return ContentList{}, j_err
	}

	werr := os.WriteFile(listFile, b, 0600)
	if werr != nil {
		return ContentList{}, werr
	}

	return newList, nil
}

func readContentList(fileListId string) (ContentList, error) {
	listfile_path := path.Join(env.Enviroment().BaseDir, LIST_PATH)

	if _, err := os.Stat(listfile_path); os.IsNotExist(err) {
		os.Mkdir(listfile_path, 0700)
	}

	file, fp_err := CreateOrOpenFile(filepath.Join(listfile_path, fileListId))
	if fp_err != nil {
		return ContentList{}, fp_err
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

func (list *ContentList) FindFile(uuid string) *ListEntry {
	for _, entry := range list.Entries {
		if entry.Id == uuid {
			return &entry
		}
	}
	return nil
}

type ContentList struct {
	Entries     []ListEntry `json:"entries"`
	StorageDir  string      `json:"base_dir"`
	Id          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
}

type ListEntry struct {
	Path     string   `json:"path"`
	FileName string   `json:"file_name"`
	Size     int64    `json:"size"`
	Sig      string   `json:"sig"`
	Type     FileType `json:"type"`
	Id       string   `json:"id"`
	ParentId string   `json:"parent_id"`
	Link     string   `json:"link"`
}
