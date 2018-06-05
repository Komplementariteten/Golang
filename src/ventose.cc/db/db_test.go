package main

import (
	"bytes"
	"os"
	"testing"
	"ventose.cc/tools"
)

type TestDocument struct {
	id         []byte
	TestString string
}

func (t *TestDocument) TypeName() string {
	return "TestDocument"
}

func BenchmarkRawReadWrite(t *testing.B) {
	cfg := &DbConfiguration{}
	cfg.Mode = 0660
	cfg.Path = "./bench.db"
	s, err := NewStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		s.Close()
		os.Remove(cfg.Path)
	}()
	s.Serve()
	//r := ConnectToStorage(c)

}

func TestClient_Relate(t *testing.T) {
	cfg := &DbConfiguration{}
	cfg.Mode = 0660
	cfg.Path = "./relation.db"
	s, err := NewStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Serve()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		s.Close()
		os.Remove(cfg.Path)
	}()
	client, err := NewClient()
	defer client.Close()
	if err != nil {
		t.Fatal(err)
	}
	d := &TestDocument{}
	rs := tools.GetRandomAsciiString(12)
	d.TestString = rs
	md, err := client.Add("TestDocument", d)
	if err != nil {
		t.Fatal(err)
	}
	d2 := &TestDocument{}
	rs = tools.GetRandomAsciiString(12)
	d.TestString = rs
	md2, err := md.Link(d2.TypeName(), d2)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(md2.relationId, md.relationId) {
		t.Fatalf("raltion IDs between origin and relation don#t match")
	}
}

func TestNewClient(t *testing.T) {
	cfg := &DbConfiguration{}
	cfg.Mode = 0660
	cfg.Path = "./client.db"
	s, err := NewStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Serve()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		s.Close()
		os.Remove(cfg.Path)
	}()
	client, err := NewClient()
	defer client.Close()
	if err != nil {
		t.Fatal(err)
	}
	d := &TestDocument{}
	rs := tools.GetRandomAsciiString(12)
	d.TestString = rs

	md, err := client.Add("TestDocument", d)
	if err != nil {
		t.Fatal(err)
	}

	q := &Query{}
	q.ID = md.Id
	q.DataType = d.TypeName()
	q.Type = READ

	err = client.Query(q).Then(func(document Document) error {
		t.Logf("%v - %v", document.Id, document.Value)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	d.TestString = tools.GetRandomAsciiString(512)
	dn := &Document{TypeName: d.TypeName(), Value: d, Id: md.Id}
	md2, err2 := client.Update(dn)
	if err2 != nil {
		t.Fatal(err2)
	}
	if !bytes.Equal(md2.Id, md.Id) {
		t.Fatalf("Update returned new Id")
	}

	tdl := &TestDocument{}
	tdl.TestString = tools.GetRandomAsciiString(128)
	md2, err = md2.Link(tdl.TypeName(), tdl)
	if err != nil {
		t.Fatal(err)
	}

	if len(md2.relationId) == 0 {
		t.Fatalf("Link returned no valid ManagedDocument, relationId is missing")
	}

	if len(md2.related) != 1 {
		t.Fatalf("Managed Document does not contain the reverse Relation")
	}

}

/* func TestConnectToStorage(t *testing.T) {
	// Should Fail without new Storage
	defer func() {
		t.Log("recovering from panic")
		recover()
	}()
	c := make(chan *Query)
	ConnectToStorage(c)

	cfg := &DbConfiguration{}
	cfg.Mode = 0660
	cfg.Path = "./test2.db"
	s, err := NewStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	s.Serve()
	defer func () {
		os.Remove(cfg.Path)
		s.Close()
	} ()
	r := ConnectToStorage(c)
	if r == nil {
		t.Fatal("Connect To Storage returend no response Channel")
	}
	d := &TestDocument{}
	rs := tools.GetRandomAsciiString(12)
	doc := &Document{}
	doc.Value = d
	doc.TypeName = "TestDocument"
	d.TestString = rs

	q := &Query{}
	q.Payload = doc
	q.Type = WRITE
	q.DataType = doc.TypeName

	c <- q
	resp := <- r
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}
	close(c)
}*/

func TestStorageRawReadWrite(t *testing.T) {
	cfg := &DbConfiguration{}
	cfg.Mode = 0660
	cfg.Path = "./test.db"
	s, err := NewStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Remove(cfg.Path)
		s.Close()
	}()
	s.Serve()
	d := &TestDocument{}
	rs := tools.GetRandomAsciiString(12)
	doc := &Document{}
	doc.Value = d
	doc.TypeName = "TestDocument"
	d.TestString = rs
	s.Writer <- doc
	resp := <-s.Result
	if resp.Status == ERROR {
		t.Fatal(resp.Error)
	}

	q := &Query{}
	q.ID = resp.Key
	q.Type = READ
	q.DataType = doc.TypeName
	s.Reader <- q
	r2 := <-s.Result
	if r2.Status == ERROR {
		t.Fatal(r2.Error)
	}
	t.Logf("%t(%v)", r2, r2)
}

func TestNewStorage(t *testing.T) {
	cfg := &DbConfiguration{}
	s, err := NewStorage(cfg)
	if err == nil {
		t.Fatal("NewStorage returns no Error with empty Configuration")
	}
	defer s.Close()

	cfg.Mode = 0660
	s, err = NewStorage(cfg)
	if err == nil {
		t.Fatal("NewStorage should Error without proper File Name")
	}
	defer s.Close()

	cfg.Path = "./sampledb.db"
	s, err = NewStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		s.Close()
		os.Remove(cfg.Path)
	}()
}
