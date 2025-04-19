package file_list

import (
	"FileBrowser/common"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

func FileListHandler(w http.ResponseWriter, r *http.Request) {
	s, serr := common.SessionFromCookie(r)
	if serr != nil {
		log.Println(serr)
	}
	if s == nil {
		http.Error(w, "Can't start Session", http.StatusBadRequest)
		return
	}

	fileListId := strings.Replace(r.URL.Path, common.LIST_PATH, "", 1)
	if fileListId != s.ContentListId {
		http.NotFound(w, r)
		return
	}

	l, rerr := readFileList(s.ContentListId)
	if rerr != nil {
		http.Error(w, rerr.Error(), http.StatusBadRequest)
	}
	t, te := template.ParseFiles("static/base.html", "file_list/file_list.html")
	if te != nil {
		panic(te)
	}
	p := &common.Page{
		Title:   "File List",
		Content: l,
	}
	err := t.Execute(w, p)
	if err != nil {
		return
	}
}

func readFileList(path string) (*FolderContent, error) {

	m, merr := fs.Glob(os.DirFS(path), "*")
	if merr != nil {
		return &FolderContent{}, merr
	}

	var items []FileItem
	var folders []string
	for _, f := range m {
		fileInfo, fierr := os.Stat(f)
		if fierr != nil {
			return &FolderContent{}, fierr
		}
		if fileInfo.IsDir() {
			folders = append(folders, f)
			continue
		}

		items = append(items, FileItem{
			Name: f,
			Size: fileInfo.Size(),
		})
	}
	return &FolderContent{
		Files: items,
		Cwd:   path,
	}, nil
}

type FolderContent struct {
	Cwd        string
	FolderPath []string
	Files      []FileItem
}

type FileItem struct {
	Name string
	Size int64
	Code string
}
