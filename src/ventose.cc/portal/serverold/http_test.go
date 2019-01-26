package serverold

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func getUrl(url string, testdata []byte) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if len(testdata) > 0 {
		d, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		if !bytes.Equal(d, testdata) {
			return fmt.Errorf("Repsonse Data is not Like testdata %s != %s", testdata, d)
		}

	}
	return nil
}

func StartHttpFrontend(cfg *HttpFrontendConfiguration) error {
	s, err := NewHttpFrontend(cfg)
	if err != nil {
		return err
	}
	go func() {
		fmt.Println("S")
		s.Srv.Serve(s.Ln)
		fmt.Println("E")
	}()
	return nil
}

func startServer(t *testing.T) {

	cfg := new(HttpFrontendConfiguration)
	os.Mkdir("/tmp/static", 0770)
	cfg.PublicStaticFilesDir = "/tmp/static/"
	os.Mkdir("/tmp/files", 0770)
	cfg.UploadDir = "/tmp/files/"
	cfg.Host = "localhost"
	cfg.Port = 8278
	e := StartHttpFrontend(cfg)
	if e != nil {
		t.Fatal(e)
	}
}

func testStaticFile(t *testing.T) {
	d := []byte("Hallo Welt")
	err := ioutil.WriteFile("/tmp/static/test.html", d, 0666)
	if err != nil {
		t.Fatal(err)
	}
	err = getUrl("http://localhost:8278/res/test.html", d)

	if err != nil {
		t.Fatal(err)
	}

}

func TestStartHttpFrontend(t *testing.T) {
	startServer(t)
	fmt.Println("t")
	testStaticFile(t)
}
