package storage

import (
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
	"os"
	"time"
)

type Store struct {
	Db               *bolt.DB
	Meta             *bolt.DB
	typeRegistration chan *MetaOpperation
}

type StoreParameter struct {
	DbFilePath string
	Timeout    time.Duration
	Mode       os.FileMode
}

type StorableType interface {
	Name() []byte
	Serialize() []byte
	Fields() []string
}

func NewStore(p *StoreParameter) (*Store, error) {
	s := &Store{}

	db, err := bolt.Open(p.DbFilePath, p.Mode, &bolt.Options{Timeout: p.Timeout})
	if err != nil {
		return nil, err
	}
	s.Db = db

	metadb, err := bolt.Open(p.DbFilePath+".meta", p.Mode, &bolt.Options{Timeout: p.Timeout})
	if err != nil {
		return nil, err
	}
	s.Meta = metadb
	go s.handleTypeRegistrations()
	return s, nil
}

func (s *Store) Close() {
	if s.typeRegistration != nil {
		close(s.typeRegistration)
		s.typeRegistration = nil
	}
	s.Meta.Close()
	s.Db.Close()
}

func appendBytes(a []byte, b []byte) []byte {
	tl := len(a) + len(b) + 4
	zeros := make([]byte, 4)
	fmt.Printf("Zeros: %v", zeros)
	r := make([]byte, tl)
	var i int
	i += copy(r[i:], a)
	i += copy(r[i:], zeros)
	copy(r[i:], b)
	return r
}

func splitBytes(v []byte) [][]byte {
	zeros := make([]byte, 4)
	return bytes.Split(v, zeros)
}
