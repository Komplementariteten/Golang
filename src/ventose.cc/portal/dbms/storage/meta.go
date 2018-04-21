package storage

import (
	"fmt"
	"github.com/boltdb/bolt"
	"bytes"
	"encoding/gob"
	"errors"
)


type MetaOpperationType uint8

const (
	METABUCKET_TYPES = "typebucket"
	MAXSEARCHVALUESIZE = 1024 * 1024 * 8
	METADESC_FILED = "description"
	METAFILEDS_FIELD = "fields"
	METAQUEUESIZE = 128
	METAOPP_CREATE MetaOpperationType = iota
	METAOPP_DELETE
	METAOPP_UPDATE
	METAOPP_FIELDLIST
	METAOPP_ADDSEARCHVALUE
)

type MetaOpperation struct {
	Opp MetaOpperationType
	Payload StorableType
}

type MetaResponse struct {
	Payload []byte
}

type AddSearchValue interface {
	TypeName() []byte
	FieldName() []byte
	Value() []byte
	Owner() []byte
}

func (s *Store) handleTypeRegistrations() {
	s.typeRegistration = make(chan *MetaOpperation, METAQUEUESIZE)
	for {
		opp, ok := <- s.typeRegistration
		if ok {
			s.handleMetaAdministration(opp)
		} else {
			return
		}
	}
}

func (s *Store) handleMetaSearch(opp AddSearchValue) (*MetaResponse,error) {

	if len(opp.Value()) > MAXSEARCHVALUESIZE || len(opp.FieldName()) > MAXSEARCHVALUESIZE || len(opp.Owner()) > MAXSEARCHVALUESIZE || len(opp.TypeName()) > MAXSEARCHVALUESIZE {
		return nil, errors.New(fmt.Sprintf("%d is to big, searchable values can not extend %d", len(opp.Value()), MAXSEARCHVALUESIZE))
	}
	response := &MetaResponse{}
	e := s.Db.Update(func(tx *bolt.Tx) error {
		metabucket := tx.Bucket([]byte(METABUCKET_TYPES))
		if metabucket == nil {
			return errors.New("Metabucket hasn't been created yet")
		}
		typebucket := metabucket.Bucket(opp.TypeName())
		if typebucket == nil {
			return errors.New(string(opp.TypeName()) + " has not been created yet")
		}

		tl := len(opp.FieldName()) + len(opp.Value()) + 1

		key := make([]byte, tl)
		var i int
		i += copy(key[i:], opp.FieldName())
		i += copy(key[i:], []byte("="))
		copy(key[i:], opp.Value())

		v := typebucket.Get(key)
		owner := opp.Owner()
		if v != nil {
			owner = appendBytes(v, owner)
		}
		typebucket.Put(key, owner)
		return nil
	})
	return response, e
}

func (s *Store) handleMetaAdministration(opp *MetaOpperation) (*MetaResponse, error) {
	var buff bytes.Buffer
	response := &MetaResponse{}
	e := s.Db.Update(func(tx *bolt.Tx) error {
		metabucket, err := tx.CreateBucketIfNotExists([]byte(METABUCKET_TYPES))
		if err != nil {
			return err
		}
		switch opp.Opp {
		case METAOPP_FIELDLIST:
			tbucket := metabucket.Bucket(opp.Payload.Name())
			if tbucket == nil {
				return errors.New(fmt.Sprintf("%s could not be opend", opp.Payload.Name()))
			}
			response.Payload = tbucket.Get([]byte(METAFILEDS_FIELD))
		case METAOPP_UPDATE:
			tbucket := metabucket.Bucket(opp.Payload.Name())
			if tbucket == nil {
				return errors.New(fmt.Sprintf("%s could not be opend", opp.Payload.Name()))
			}
			tbucket.Put([]byte(METADESC_FILED), opp.Payload.Serialize())
			enc := gob.NewEncoder(&buff)
			enc.Encode(opp.Payload.Fields())
			tbucket.Put([]byte(METAFILEDS_FIELD), buff.Bytes())
			buff.Reset()
		case METAOPP_CREATE:
			tbucket, err := metabucket.CreateBucket(opp.Payload.Name())
			if err != nil {
				return err
			}
			tbucket.Put([]byte(METADESC_FILED), opp.Payload.Serialize())
			enc := gob.NewEncoder(&buff)
			enc.Encode(opp.Payload.Fields())
			tbucket.Put([]byte(METAFILEDS_FIELD), buff.Bytes())
			buff.Reset()
		case METAOPP_DELETE:
			err := metabucket.DeleteBucket(opp.Payload.Name())
			if err != nil {
				return err
			}
		}
		return nil
	})
	return response, e
}
