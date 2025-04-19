package main

import (
	"FileBrowser/common"
	"FileBrowser/file_list"
	"FileBrowser/pki"
	"FileBrowser/session"
	"log"
	"net/http"
)

func main() {
	pki.Init()
	fs := http.FileServer(http.Dir("static"))
	http.Handle(common.STATIC_PATH, http.StripPrefix(common.STATIC_PATH, fs))
	http.HandleFunc(common.LIST_PATH, file_list.FileListHandler)
	http.HandleFunc(common.SESSION_PATH, session.SessionHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
