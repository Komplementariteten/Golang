package main

import (
	"bytes"
	"fmt"
	"os"
	"ventose.cc/tools"
)

type TestDocument struct {
	id         []byte
	TestString string
}

func (t *TestDocument) TypeName() string {
	return "TestDocument"
}
func main() {
	cfg := &DbConfiguration{}
	cfg.Mode = 0660
	cfg.Path = "./client.db"
	s, err := NewStorage(cfg)
	if err != nil {
		panic(err)
	}
	err = s.Serve()
	if err != nil {
		panic(err)
	}
	defer func() {
		s.Close()
		os.Remove(cfg.Path)
	}()
	client, err := NewClient()
	defer client.Close()
	if err != nil {
		panic(err)
	}
	d := &TestDocument{}
	rs := tools.GetRandomAsciiString(12)
	d.TestString = rs

	md, err := client.Add("TestDocument", d)
	if err != nil {
		panic(err)
	}

	q := &Query{}
	q.ID = md.Id
	q.DataType = d.TypeName()
	q.Type = READ

	err = client.Query(q).Then(func(document Document) error {
		fmt.Printf("Query OK %v - %v \n", document.Id, document.Value)
		return nil
	})
	if err != nil {
		panic(err)
	}

	d.TestString = tools.GetRandomAsciiString(512)
	dn := &Document{TypeName: d.TypeName(), Value: d, Id: md.Id}
	md2, err2 := client.Update(dn)
	if err2 != nil {
		panic(err2)
	}
	if !bytes.Equal(md2.Id, md.Id) {
		panic("Update returned new Id")
	}

	tdl := &TestDocument{}
	tdl.TestString = tools.GetRandomAsciiString(128)
	md2, err = md2.Link(tdl.TypeName(), tdl)
	if err != nil {
		panic(err)
	}

	if len(md2.relationId) == 0 {
		panic("Link returned no valid ManagedDocument, relationId is missing")
	}

	if len(md2.related) != 1 {
		panic("Managed Document does not contain the reverse Relation")
	}
	fmt.Println("done")
}
