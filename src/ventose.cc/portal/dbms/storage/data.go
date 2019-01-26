package storage

import (
	"fmt"
	"github.com/boltdb/bolt"
	"ventose.cc/tools"
)

type QueryType uint8

const (
	QUERYTYPE_READ QueryType = iota
	QUERYTYPE_WRITE
	QUERYTYPE_ADD
	QUERYTYPE_DELETE
	MINBUCKETNAME_SIZE = 1
	MINKEY_SIZE        = 32
)

var packageList map[string]*Store

type Document struct {
	Key     []byte
	Payload StorableType
}

type Query struct {
	Bucket  string
	Type    QueryType
	Content *Document
}

type Response struct {
	Error error
	Data  []byte
	Key   []byte
}

func updateDocument(bucket *bolt.Bucket, d *Document) (r *Response, err error) {

	r = new(Response)

	if len(d.Key) < MINKEY_SIZE {
		err = fmt.Errorf("%q is no valid Key", d.Key)
		return
	}

	err = bucket.Put(d.Key, d.Payload.Serialize())

	r.Key = d.Key

	return
}
func deleteDocument(bucket *bolt.Bucket, d *Document) (r *Response, err error) {

	r = new(Response)

	if len(d.Key) < MINKEY_SIZE {
		err = fmt.Errorf("%q is no valid Key", d.Key)
		return
	}

	err = bucket.Delete(d.Key)
	r.Key = d.Key

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

	err = bucket.Put(d.Key, d.Payload.Serialize())

	r.Key = d.Key

	return
}

func (s *Store) handleStorageOpperation(query *Query) (response *Response, e error) {
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
				case QUERYTYPE_ADD:
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
