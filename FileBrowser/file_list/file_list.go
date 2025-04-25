package file_list

import (
	"FileBrowser/common"
	"html/template"
	"log"
	"net/http"
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

	l := common.GetContentList(s)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
