package main

import (
	"FileBrowser/common"
	"FileBrowser/env"
	"FileBrowser/file"
	"FileBrowser/file_list"
	"FileBrowser/pki"
	"FileBrowser/session"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		wd, _ := os.Getwd()
		env.Enviroment().SetWd(wd)
	} else {
		env.Enviroment().SetWd(os.Args[1])
	}
	pki.Init()
	fs := http.FileServer(http.Dir("static"))
	http.Handle(common.STATIC_PATH, http.StripPrefix(common.STATIC_PATH, fs))
	http.HandleFunc(common.LIST_PATH, file_list.FileListHandler)
	http.HandleFunc(common.SESSION_PATH, session.SessionHandler)
	http.HandleFunc(common.FILE_PATH, file.FileHandle)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
