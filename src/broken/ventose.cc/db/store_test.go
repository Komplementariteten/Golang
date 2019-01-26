package db

import (
	"testing"
	"os"
	"bytes"
	"encoding/gob"
	"ventose.cc/tools"
	"math/big"
	"crypto/rand"
	"fmt"
)

type dataFnc func() []byte

type TestType struct {
	Name string
	Blurp []byte
	Id *big.Int
}

func descodeTestType(d []byte) *TestType {
	var buff bytes.Buffer
	tt := new(TestType)
	buff.Read(d)
	dec := gob.NewDecoder(&buff)
	dec.Decode(&tt)
	return tt
}

func getTestMap() []byte {
	d := make(map[string]string)

	for i := 0; i < 1000; i++ {
		d[tools.GetRandomAsciiString(16)] = tools.GetRandomAsciiString(64)
	}

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)

	enc.Encode(d)

	return buff.Bytes()
}

func getTestStruct() []byte {

	d := new(TestType)
	d.Name = string(tools.GetRandomBytes(12))
	d.Blurp = tools.GetRandomBytes(1024*1024*1024)
	rand.Int(rand.Reader, new(big.Int).SetUint64(1024*1024*1024*1024))
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)

	enc.Encode(d)

	return buff.Bytes()

}

func newDocument(fnc dataFnc) *Document {
	d := &Document{}
	data := fnc()
	d.Payload = data
	return d
}


func TestWriteStore(t *testing.T) {


	p := &Parameter{}
	p.Timeout = 1
	p.DbFilePath = tools.GetRandomAsciiString(6) + ".db"
	p.Mode = 0644
	defer os.Remove(p.DbFilePath)

	db, err := NewStore(p)
	defer db.Close()

	if err != nil {
		t.Error(err)
	}

	queryChan := make(chan *Query)
	ctrlChan := make(chan bool)
	respChan := db.Connect(queryChan, ctrlChan)

	q1 := &Query{}
	q1.Type = QUERYTYPE_CREATE
	q1.Content = newDocument(getTestMap)
	q1.Bucket = tools.GetRandomAsciiString(MINBUCKETNAME_SIZE + 1)
	queryChan <- q1
	r1 := <- respChan

	if r1.Error != nil {
		t.Fatal(r1.Error)
	}

	if len(r1.Key) < MINKEY_SIZE {
		t.Fatal("New Key is to small")
	}

	q2 := &Query{}
	q2.Type = QUERYTYPE_READ
	q2.Bucket = q1.Bucket
	q2.Content = new(Document)
	q2.Content.Key = r1.Key

	queryChan <- q2
	r2 := <- respChan

	if r2.Error !=  nil {
		t.Fatal(r2.Error)
	}

	if !bytes.Equal(r1.Key, r2.Key) {
		t.Error("Keys written and read do not match: %q != %q", r1.Key, r2.Key)
	}

	q3 := &Query{}

	q3.Type = QUERYTYPE_READ
	q3.Content = new(Document)
	q3.Content.Key = r1.Key
	q3.Bucket = q2.Bucket
	queryChan <- q3
	r2 = <- respChan

	if r2.Error !=  nil {
		t.Fatal(r2.Error)
	}

	if !bytes.Equal(r1.Key, r2.Key) {
		t.Error("Keys written and read do not match: %q != %q", r1.Key, r2.Key)
	}

	tt := descodeTestType(r2.Data)

	if len(tt.Name) != 12 {
		fmt.Printf("Name of wrong %v", tt)
	}

	q3.Type = QUERYTYPE_WRITE
	q3.Content = newDocument(getTestStruct)
	q3.Content.Key = r1.Key
	q3.Bucket = q2.Bucket
	queryChan <- q3
	r2 = <- respChan

	if r2.Error !=  nil {
		t.Fatal(r2.Error)
	}

	if !bytes.Equal(r1.Key, r2.Key) {
		t.Error("Keys written and read do not match: %q != %q", r1.Key, r2.Key)
	}


	q3.Type = QUERYTYPE_DELETE
	q3.Content = &Document{}
	q3.Bucket = q2.Bucket
	q3.Content.Key = r2.Key
	queryChan <- q3
	r2 = <- respChan

	if r2.Error !=  nil {
		t.Fatal(r2.Error)
	}

	queryChan <- q3
	r2 = <- respChan

	if r2.Error !=  nil {
		t.Fatal(r2.Error)
	}

	if !bytes.Equal(r1.Key, r2.Key) {
		t.Error("Keys2 written and read do not match: %q != %q", r1.Key, r2.Key)
	}


}

func TestNewStore2(t *testing.T) {
	p := &Parameter{}
	p.Timeout = 1
	p.DbFilePath = "NoRealTestFile.db"
	p.Mode = 0600

	db, err := NewStore(p)
	if err != nil {
		t.Error(err)
	}

	db2, err := NewStore(p)
	if err != nil {
		t.Error(err)
	}

	if db != db2 {
		t.Error("Newstore with same Parameter return diffrent Type")
	}

}

func TestNewStore(t *testing.T) {

	p := &Parameter{}
	p.Timeout = 1
	p.DbFilePath = "NoRealTestFile.db"
	p.Mode = 0600

	db, err := NewStore(p)
	if err != nil {
		t.Error(err)
	}
	if db.Db.IsReadOnly() {
		t.Error("Database is not writeable")
	}

	if _, err := os.Stat(p.DbFilePath); os.IsNotExist(err) {
		t.Error("Database file does not exist after closing Database")
	}
}
