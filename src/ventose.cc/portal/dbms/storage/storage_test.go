package storage

import (
	"testing"
	"ventose.cc/tools"
	"github.com/boltdb/bolt"
	"os"
)

type MyStorableType struct {
	StorableType
	myname []byte
}

func (m *MyStorableType) Name() []byte {
	return m.myname
}

func (m *MyStorableType) Serialize() []byte {
	zeroes := make([]byte, 16)
	return zeroes
}

func (m *MyStorableType) Fields() []string {
	names := []string {"abc", "def", "ghi", "jkl"}
	return names
}

func GetTestDbParams() *StoreParameter {
	p := &StoreParameter{}
	p.Timeout = 1
	p.DbFilePath = tools.GetRandomAsciiString(6) + ".db"
	p.Mode = 0600
	return p
}

func TestBucketExists(t *testing.T) {
	dbfile := "testbe.db"
	db, err := bolt.Open(dbfile, 0600 , nil )
	if err != nil {
		t.Fatal(err)
	}
	defer func (db *bolt.DB) {
		db.Close()
		os.Remove(dbfile)
	}(db)

	e := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("abcdefgij"))
		if bucket != nil {
			t.Fatal("Bucket does not exist but not nil returned!")
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}

func TestStorehandleMetaAdministration(t *testing.T) {
	s, err := NewStore(GetTestDbParams())
	defer s.Close()
	if err != nil {
		t.Fatal(err)
	}
	tt := &MyStorableType{}
	tt.myname = []byte("myname")

	opp := &MetaOpperation{ Payload: tt, Opp: METAOPP_CREATE}
	_, err = s.handleMetaAdministration(opp)
	if err != nil {
		t.Fatal(err)
	}
	opp.Opp = METAOPP_DELETE
	_, err = s.handleMetaAdministration(opp)
	if err != nil {
		t.Fatal(err)
	}

}

func TestNewStore(t *testing.T) {
	p := GetTestDbParams()
	s, err := NewStore(p)
	if err != nil {
		t.Fatal(err)
	}
	defer func () {
		os.Remove(p.DbFilePath)
	}()
	s.Close()
	//_, ok := <- s.typeRegistration
	if s.typeRegistration != nil {
		t.Fatal("typeRegistragtion not closed after Closing Store")
	}
}