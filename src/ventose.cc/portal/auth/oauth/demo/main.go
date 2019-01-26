package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"ventose.cc/auth/oauth"
	"ventose.cc/tools"
)

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)

	p, _ := os.Getwd()
	fmt.Println(p)

	//s := oauth.NewServer()
	//s.SetLoginTemplate("src/ventose.cc/auth/oauth/demo/logintemplate.html")
	var client_url = "http://www.example.org"

	c, err := oauth.NewClient("test_client", client_url)
	if err != nil {
		panic(err)
	}
	fmt.Println(c.ClientId)
	fmt.Println(tools.GetRandomAsciiString(10))
	handle := oauth.GetHandler()
	handle.AuthEndpoint.SetLoginTemplateFile("src/ventose.cc/auth/oauth/demo/logintemplate.html")
	handle.AddClient(c)
	http.Handle("/oauth/", handle)
	http.ListenAndServe(":3030", nil)

}
