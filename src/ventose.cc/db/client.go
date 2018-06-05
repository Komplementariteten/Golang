package main

import (
	"bytes"
	"fmt"
	"sync"
)

const RELATIONS = "rxds"

type Client struct {
	query    chan *Query
	response <-chan *Response
	lock     sync.Mutex
}

type Future struct {
	C *Client
}

type ManagedDocument struct {
	Id         []byte
	Type       string
	relationId []byte
	related    []*ManagedDocument
	context    *Client
}

func NewClient() (*Client, error) {
	c := &Client{}
	c.query = make(chan *Query)
	c.response = ConnectToStorage(c.query)
	return c, nil
}

func (c *Client) Close() {
	q := &Query{}
	q.Type = CLOSE
	c.query <- q
}

func (c *Client) Add(typeName string, d interface{}) (*ManagedDocument, error) {
	doc := &Document{}
	doc.Value = d
	doc.TypeName = typeName
	id, err := c.rawAdd(doc)
	if err != nil {
		return nil, err
	}
	mdoc := &ManagedDocument{Id: id, Type: typeName, context: c}
	return mdoc, nil
}

func (c *Client) Update(d *Document) (*ManagedDocument, error) {
	id, err := c.rawAdd(d)
	if err != nil {
		return nil, err
	}
	mdoc := &ManagedDocument{Id: id, Type: d.TypeName, context: c}
	return mdoc, nil
}

func (c *Client) rawAdd(d *Document) ([]byte, error) {
	q := &Query{}
	q.Payload = d
	q.Type = WRITE
	q.DataType = d.TypeName
	if len(d.Id) != 0 {
		q.ID = d.Id
	}
	c.lock.Lock()
	c.query <- q
	resp := <-c.response
	if resp.Status == ERROR {
		c.lock.Unlock()
		return nil, resp.Error
	}
	c.lock.Unlock()
	return resp.Key, nil
}

func (c *Client) Query(q *Query) *Future {
	c.query <- q
	f := &Future{}
	f.C = c
	return f
}

// []byte -> DataSet ID
// Document -> DataSet Value
func (f *Future) Then(fn func(Document) error) error {
	resp := <-f.C.response
	if resp.Status == ERROR {
		return resp.Error
	}

	if !bytes.Equal(resp.Key, resp.Value.Id) {
		panic(" Response Key and Document Id do not match in Client.Then!")
	}
	return fn(resp.Value)
}

func (m *ManagedDocument) Link(typeName string, d interface{}) (*ManagedDocument, error) {
	doc := &Document{}
	doc.TypeName = typeName
	doc.Value = d
	if m.relationId == nil {
		did, err := m.context.rawAdd(doc)
		if err != nil {
			return nil, err
		}
		idlist := [][]byte{did}
		rdoc := &Document{TypeName: RELATIONS, Value: idlist}
		rid, err := m.context.rawAdd(rdoc)
		if err != nil {
			return nil, err
		}
		mrdoc := &ManagedDocument{Id: did, context: m.context, Type: typeName, related: []*ManagedDocument{m}, relationId: rid}
		m.relationId = rid
		m.related = []*ManagedDocument{mrdoc}
		return m, nil
	} else {
		rq := &Query{Type: READ, DataType: RELATIONS, ID: m.relationId}
		did, err := m.context.rawAdd(doc)
		if err != nil {
			return nil, err
		}
		mrdoc := &ManagedDocument{Id: did, context: m.context, Type: typeName, related: []*ManagedDocument{m}, relationId: m.relationId}
		err = m.context.Query(rq).Then(func(doc Document) error {
			idlist, ok := doc.Value.([][]byte)
			if ok {
				idlist = append(idlist, did)
				rdoc := &Document{TypeName: RELATIONS, Id: m.relationId, Value: idlist}
				chid, err := m.context.rawAdd(rdoc)
				if err != nil {
					return err
				}
				if !bytes.Equal(chid, m.relationId) {
					return fmt.Errorf("Updateing Relation List Failed %v != %v", chid, m.relationId)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		m.related = append(m.related, mrdoc)
		return m, nil
	}
}
