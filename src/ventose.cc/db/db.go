package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"os"
	"sync"
	"ventose.cc/tools"
)

const (
	BUFF_SIZE                        = 64
	COMMAND_TYPES                    = 3
	RESPONSE_TYPES                   = 5
	MIN_BUCKETNAME_SIZE              = 3
	QUERY_TYPES         QueryType    = 7
	CLOSE_COMMAND                    = iota * COMMAND_TYPES
	ERROR               ResponseType = iota * RESPONSE_TYPES
	UPDATE
	CREATE
	NOTFOUND
	READ QueryType = iota * QUERY_TYPES
	WRITE
	DELETE
	CLOSE
)

type QueryType int
type ResponseType int

type Document struct {
	TypeName string
	Value    interface{}
	Id       []byte
}

type Storage struct {
	hndl    *bolt.DB
	Reader  chan *Query
	Result  chan *Response
	Writer  chan *Document
	Control chan *Command
	// Cursor chan chan *Document
	Clients sync.WaitGroup
	Lock    sync.Mutex
}

type Response struct {
	Status ResponseType
	Error  error
	Value  Document
	Key    []byte
}

type Command struct {
}

type Query struct {
	DataType string
	ID       []byte
	Type     QueryType
	Payload  *Document
}

type DbConfiguration struct {
	Path string
	Mode os.FileMode
	Opt  *bolt.Options
}

var mStorage *Storage = nil

/*
	Creates new BoltDB Backend Type with initialized Channels
*/
func NewStorage(cfg *DbConfiguration) (s *Storage, err error) {

	if cfg.Mode <= 0600 {
		return nil, fmt.Errorf("FileMode %d has no valid value", cfg.Mode)
	}
	fi, e := os.Stat(cfg.Path)
	if os.IsPermission(e) {
		return nil, e
	} else if os.IsExist(e) {
		return nil, fmt.Errorf("The File %s already exists: %v", cfg.Path, e)
	} else if e != nil && !os.IsNotExist(e) {
		return nil, fmt.Errorf("Can't create %s: %v", cfg.Path, e)
	} else if fi != nil {
		return nil, fmt.Errorf("%s Exists", cfg.Path)
	}

	h, e := os.OpenFile(cfg.Path, os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return nil, fmt.Errorf("%s is not createable", cfg.Path)
	}
	defer h.Close()
	e = os.Remove(cfg.Path)
	if e != nil {
		return nil, e
	}

	s = &Storage{}

	if cfg.Opt != nil {
		s.hndl, err = bolt.Open(cfg.Path, cfg.Mode, nil)
	} else {
		s.hndl, err = bolt.Open(cfg.Path, cfg.Mode, cfg.Opt)
	}
	if err != nil {
		return
	}
	s.Reader = make(chan *Query)
	s.Result = make(chan *Response, BUFF_SIZE)
	s.Writer = make(chan *Document)
	s.Control = make(chan *Command)
	mStorage = s
	return
}

/*
	Closes Current BoltDB
*/
func (s *Storage) Close() {
	if s != nil {
		s.Lock.Lock()
		s.hndl.Close()
		s.Clients.Wait()
		close(s.Reader)
		close(s.Writer)
		close(s.Control)
		close(s.Result)
	}
}

/*
Connect a Client to this Storage
*/
func ConnectToStorage(ch chan *Query) <-chan *Response {

	if mStorage == nil {
		panic("no Storage initialized")
	}

	r := make(chan *Response)
	go func() {
		mStorage.Clients.Add(1)
		for q := range ch {
			if q.Type == READ {
				mStorage.Reader <- q
				resp := <-mStorage.Result
				r <- resp
			} else if q.Type == WRITE {
				mStorage.Writer <- q.Payload
				resp := <-mStorage.Result
				r <- resp
			} else if q.Type == CLOSE {
				break
			} else {
				resp := &Response{}
				resp.Status = NOTFOUND
				resp.Error = fmt.Errorf("Not Implemented!")
			}
		}
		mStorage.Clients.Done()
	}()
	return r
}

func (s *Storage) read(q *Query) *Response {
	r := &Response{}
	if len(q.DataType) <= MIN_BUCKETNAME_SIZE {
		r.Status = ERROR
		r.Error = fmt.Errorf("DataType must contain at least %d Characters len(%s)=%d", MIN_BUCKETNAME_SIZE, q.DataType, len(q.DataType))
		return r
	}
	err := s.hndl.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(q.DataType))
		v := b.Get(q.ID)
		if v != nil {
			vt := &Document{}
			tools.Decode(v, &vt)
			if len(vt.Id) == 0 {
				vt.Id = q.ID
			}
			r.Value = *vt
			r.Key = q.ID
		} else {
			r.Status = NOTFOUND
			r.Error = fmt.Errorf("Key:%v not found", q.ID)
		}
		return nil
	})
	if err != nil {
		r.Status = ERROR
		r.Error = err
	}
	return r
}

func (s *Storage) write(q *Document) *Response {
	r := &Response{}
	if len(q.TypeName) <= MIN_BUCKETNAME_SIZE {
		r.Status = ERROR
		r.Error = fmt.Errorf("DataType must contain at least %d Characters len(%s)=%d", MIN_BUCKETNAME_SIZE, q.TypeName, len(q.TypeName))
		return r
	}

	err := s.hndl.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(q.TypeName))
		if b == nil {
			buck, err := tx.CreateBucket([]byte(q.TypeName))
			if err != nil {
				return err
			}
			b = buck
		}
		id := q.Id
		if len(id) == 0 {
			uid, e := b.NextSequence()
			if e != nil {
				return e
			}
			id = tools.Itob(int(uid))
			r.Status = CREATE
		} else {
			r.Status = UPDATE
		}
		by := tools.EncodeToBytes(q)
		r.Key = id
		return b.Put(id, by)
	})
	if err != nil {
		r.Status = ERROR
		r.Error = err
	}
	return r
}

func (s *Storage) Serve() error {
	go func() {
		for {
			select {
			case cmd := <-s.Control:
				s.Lock.Lock()
				cls := s.handleCommand(cmd)
				if cls != nil {
					s.Result <- cls
				}
				s.Lock.Unlock()
			case q := <-s.Reader:
				s.Lock.Lock()
				response := s.read(q)
				s.Result <- response
				s.Lock.Unlock()
			case d := <-s.Writer:
				s.Lock.Lock()
				doc := s.write(d)
				s.Result <- doc
				s.Lock.Unlock()
			}
		}
	}()
	return nil
}
