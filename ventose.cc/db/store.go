package db

import (
	"fmt"
	"time"
	"github.com/boltdb/bolt"
	"os"
	"log"
	"ventose.cc/tools"
)

type QueryType uint8

const (
	QUERYTYPE_READ QueryType = iota
	QUERYTYPE_WRITE
	QUERYTYPE_CREATE
	QUERYTYPE_DELETE
	MINBUCKETNAME_SIZE = 1
	MINKEY_SIZE = 32
)


var packageList map[string]*Store

type Document struct {
	Key []byte
	Payload []byte
}
type Query struct {
	Bucket string
	Type QueryType
	Content *Document
}

type Response struct {
	Error error
	Data []byte
	Key []byte
}

type Store struct {
	Db *bolt.DB
	repChans []chan *Response
}



type Parameter struct {
	DbFilePath string
	Timeout time.Duration
	Mode os.FileMode
}

func NewStore(p *Parameter) (*Store, error) {
	s := &Store{}

	if packageList == nil {
		packageList = make(map[string]*Store)
	}

	if _, ok := packageList[p.DbFilePath]; ok {
		return packageList[p.DbFilePath], nil
	}

	db, err := bolt.Open(p.DbFilePath,p.Mode, &bolt.Options{Timeout: p.Timeout} )
	if err != nil {
		return nil, err
	}

	s.Db = db
	packageList[p.DbFilePath] = s

	return s, nil
}

func updateDocument(bucket *bolt.Bucket, d *Document) (r *Response, err error) {

	r = new(Response)

	if len(d.Key) < MINKEY_SIZE {
		err = fmt.Errorf("%q is no valid Key", d.Key)
		return
	}

	err = bucket.Put(d.Key, d.Payload)

	r.Key = d.Key

	return
}
func deleteDocument(bucket *bolt.Bucket, d *Document) (r *Response, err error) {

	r = new(Response)

	if len(d.Key) < MINKEY_SIZE {
		err = fmt.Errorf("%q is no valid Key", d.Key)
		return
	}

	data := bucket.Get(d.Key)

	if len(data) > 0 {
		r.Data = data
		err = bucket.Delete(d.Key)
	} else {
		err = fmt.Errorf("%q not found", d.Key)
	}

	if err != nil {
		r.Error = err
	}

	return
}
func createDocument(bucket *bolt.Bucket, d *Document) (r *Response, err error) {

	r = new(Response)

	if len(d.Key) < MINKEY_SIZE {
		d.Key = tools.GetRandomBytes(MINKEY_SIZE)
	}

	err = bucket.Put(d.Key, d.Payload)

	r.Key = d.Key

	return
}

func (s *Store) handleQuery(query *Query) ( response *Response, e error) {

	if response == nil {
		response = new(Response)
	}

	if len(query.Bucket) < MINBUCKETNAME_SIZE {
		return nil, fmt.Errorf("No Bucket Found in Query, or bucket name to short")
	}

	if query.Type == QUERYTYPE_READ {

		e = s.Db.View(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(query.Bucket))
				if bucket == nil {
					return fmt.Errorf("Bucket %q not found", query.Bucket)
				}
				if len(query.Content.Key) < MINKEY_SIZE {
					return fmt.Errorf("No Key or wrong found %q", query.Content.Key)
				}
				val := bucket.Get(query.Content.Key)
				if val == nil {
					return fmt.Errorf("Failed to get Item for Key %q", query.Content.Key)
				}

				response.Data = val
				response.Key = query.Content.Key
				return nil
		})

	} else {
		e = s.Db.Update(func(tx *bolt.Tx) error {
			if len(query.Bucket) > MINBUCKETNAME_SIZE {
				bucket, err := tx.CreateBucketIfNotExists([]byte(query.Bucket))
				if err != nil {
					return err
				}
				switch query.Type {
				case QUERYTYPE_CREATE:
					response, err = createDocument(bucket, query.Content)
					return err
				case QUERYTYPE_DELETE:
					response, err = deleteDocument(bucket, query.Content)
					return err
				case QUERYTYPE_WRITE:
					response, err = updateDocument(bucket, query.Content)
					return err
				default:
					return fmt.Errorf("Query Type %q is not supported", query.Type)

				}
				return nil
			} else {
				return fmt.Errorf("No Bucket Found in Query, or bucket name to short")
			}
		})
	}


	return
}

func (s* Store) closeInt() {
	if s != nil {
		s.Db.Close()
		for channel := range s.repChans {
			close(s.repChans[channel])
		}
	}
}

func (s* Store) Connect(in <-chan *Query, quit <- chan bool) <- chan *Response {
	out := make(chan *Response)

	go func() {
		for {
			select {
			case q := <- in:
				r, err := s.handleQuery(q)

				if r == nil {
					r = new(Response)
				}

				if err != nil {
					log.Println(err.Error())
					r.Error = err
				}
				out <- r
			case <- quit:
				close(out)
				log.Println("Closing Connection")
				break
			}
		}
	}()
	return out
}

func (s* Store) Close() {
	s.closeInt()
}