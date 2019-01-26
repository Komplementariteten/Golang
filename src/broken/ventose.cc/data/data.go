package data

import (
	"fmt"
	lediscfg "github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
	"ventose.cc/tools"
)

const (
	DefaultDb = iota
	TypeDb
	MetaDb
	SearchDb
	SystemDb
	ShutdownCmd = iota
	ShutdownDone
	CreateRequest
	DeleteRequest
	ReadRequest
	ListRequest
	DescribeRequest
	UpdateRequest
	SearchRequest
	DELETED    = "deleted"
	INDEX_SIZE = 64
)

/*
Storage Type
*/
type DataStorage struct {
	l           *ledis.Ledis
	c           *ledis.DB
	scope       string
	transaction bool
	controll    chan int
	Running     bool
	mtx         sync.Mutex
	wg          sync.WaitGroup
	clients     sync.WaitGroup
	clientCount int64
}

/*
Client Connection
*/
type StorageConnection struct {
	RequestChannel  chan StorageRequest
	ResponseChannel chan StorageResponse
	storage         *DataStorage
}

/*
Meta Information on Request Items
*/
type ItemMetaInformation struct {
	Created time.Time
	Updated time.Time
	Type    []byte
}

/*
Response from Stroage Request
*/
type StorageResponse struct {
	Type     int
	Error    error
	Affected []byte
	Content  RequestContent
}

type DescribeResponse struct {
	RequestContent
	Found bool
	Count int64
	Meta  *ItemMetaInformation
}

/*
List Response
*/

type ListResponse struct {
	RequestContent
	Count    int64
	Type     []byte
	Elements [][]byte
	Content  []RequestContent
}

/*
Request Content
*/
type RequestContent interface {
}

/*
Request Parametes
*/
type ParameterType struct {
	Offset int32
	Top    int32
	SKey   string
	SValue interface{}
}

/**
Request to Storage
*/
type StorageRequest struct {
	Type      int
	Content   RequestContent
	Element   []byte
	Hash      []byte
	Parameter *ParameterType
}

/**
Ledis Configuration
*/
type InitialConfiguration struct {
	DbPath         string
	PathAccessMode os.FileMode
	MaxDatabases   int
	AuthKey        string
	ConfigPath     string
}

func (c *StorageConnection) Close() {
	c.storage.Disconnect()
}

func (d *DataStorage) Disconnect() {
	d.clients.Done()
	d.clientCount--
}

func (t *StorageRequest) Keys() []string {
	keys := []string{}
	s := reflect.ValueOf(&t.Content).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		name := typeOfT.Field(i).Name
		//s := typeOfT.Field(i).Type.Name()
		if !strings.HasPrefix(name, "_") {
			keys = append(keys, name)
		}
	}

	return keys
}

func (d *DataStorage) AddSearchKey(t interface{}, index []byte) {
	d.Switchb(SearchDb)

	val := reflect.ValueOf(t).Elem()
	for n := 0; n < val.NumField(); n++ {
		var searchKey string
		valueField := val.Field(n)
		typeField := val.Type().Field(n)
		tag := typeField.Tag.Get("search")

		if tag != "" {
			searchKey = fmt.Sprintf("%T[%s=%v]", t, tag, valueField.Interface())
			key := tools.ToBytes(searchKey)
			d.c.Set(key, index)
		}
	}
	d.Switchb(DefaultDb)
}

func (d *DataStorage) search(t interface{}) ([]byte, error) {
	d.Switchb(SearchDb)

	val := reflect.ValueOf(t).Elem()
	for n := 0; n < val.NumField(); n++ {
		var searchKey string
		valueField := val.Field(n)
		typeField := val.Type().Field(n)
		tag := typeField.Tag.Get("search")
		if tag != "" {
			searchKey = fmt.Sprintf("%T[%s=%v]", t, tag, valueField.Interface())
			key := tools.ToBytes(searchKey)
			data, err := d.c.Get(key)
			return data, err
		}
	}
	return nil, fmt.Errorf("No Data found")
}

func (d *DataStorage) Process(requests <-chan StorageRequest, con *StorageConnection) {
	for req := range requests {
		d.wg.Add(1)
		go func() {
			defer func() {

				resp := new(StorageResponse)
				if r := recover(); r != nil {
					if err, ok := r.(error); ok {
						resp.Error = err
						con.ResponseChannel <- *resp
						d.wg.Done()
					}
				}
			}()
			d.mtx.Lock()
			response := new(StorageResponse)
			switch req.Type {
			default:
				response.Error = fmt.Errorf("Unkown Option %s", req.Type)
			case DescribeRequest:
				meta := &DescribeResponse{}
				if req.Element != nil && len(req.Element) > 0 {
					d.Switchb(MetaDb)
					item, err := d.c.Get(req.Element)
					if err != nil {
						response.Error = err
						break
					}
					metaInfo := new(ItemMetaInformation)
					tools.Decode(item, metaInfo)
					meta.Meta = metaInfo
					meta.Found = true
				} else {
					d.Switchb(TypeDb)
					meta.Found = false
					ctype := req.Content
					typeString := reflect.TypeOf(ctype).String()
					ty := tools.ToBytes(typeString)
					count, err := d.c.LLen(ty)
					meta.Count = count
					if err != nil {
						fmt.Println("Error")
						response.Error = err
						break
					}

				}
				response.Content = meta
				d.Switchb(DefaultDb)
			case ListRequest:
				ctype := req.Content
				typeString := reflect.TypeOf(ctype).String()
				ty := tools.ToBytes(typeString)
				d.Switchb(TypeDb)
				meta := &ListResponse{}
				count, err := d.c.LLen(ty)
				if err != nil {
					response.Error = fmt.Errorf("[DataBacken] ListRequest error: %s", err.Error())
					break
				}
				meta.Count = count
				meta.Type = ty
				indexList := [][]byte{}
				for l := 0; l < int(count); l++ {
					elem, ierr := d.c.LIndex(ty, int32(l))
					if ierr == nil {
						indexList = append(indexList, elem)
					}
				}
				meta.Elements = indexList

				if req.Parameter == nil {
					req.Parameter = new(ParameterType)
					req.Parameter.Offset = 0
					req.Parameter.Top = int32(count)
				}

				elemList := [][]byte{}
				//pType := req.Content
				if req.Parameter.Top > 0 {
					elems, _ := d.c.LRange(ty, req.Parameter.Offset, req.Parameter.Top)
					for _, el := range elems {
						elemList = append(elemList, el)
					}
				}

				d.Switchb(DefaultDb)
				pType := req.Content
				for _, el := range elemList {
					item, err := d.c.Get(el)
					if err == nil {
						tools.Decode(item, pType)
						meta.Content = append(meta.Content, pType)
					}
				}

				response.Content = meta
			case CreateRequest:
				ctype := req.Content
				b := tools.EncodeToByte(ctype)
				index := tools.GetRandomBytes(INDEX_SIZE)
				err := d.c.Set(index, b)
				if err != nil {
					response.Error = fmt.Errorf("[DataBacken] CreateRequest error: %s", err.Error())
				} else {
					response.Affected = index
					d.Switchb(TypeDb)
					typeString := reflect.TypeOf(ctype).String()
					t := tools.ToBytes(typeString)
					d.c.LPush(t, index)
					d.c.Set(index, t)
					d.Switchb(MetaDb)
					meta := &ItemMetaInformation{time.Now(), time.Now(), t}
					metab := tools.EncodeToByte(meta)
					d.c.Set(index, metab)
					d.AddSearchKey(ctype, index)
				}
			case UpdateRequest:
				if test, err := d.c.Exists(req.Element); err != nil || test <= 0 {
					response.Error = fmt.Errorf("UpdateRequest failed cause Ledis can't Find Element with: %v/%v for %v", err, test, req)
				} else {
					//TODO Handle Type Change!
					b := tools.EncodeToByte(req.Content)
					d.c.Set(req.Element, b)
					response.Affected = req.Element
					d.Switchb(MetaDb)

					// Update Metainformation
					bin, _ := d.c.Get(req.Element)
					meta := &ItemMetaInformation{}
					tools.Decode(bin, &meta)
					meta.Updated = time.Now()
					bin = tools.EncodeToByte(meta)
					d.c.Set(req.Element, bin)

					d.Switchb(DefaultDb)
					d.AddSearchKey(req.Content, req.Element)
					//d.Switchb(TypeDb)
					//typemap, _ := d.c.LPop(req.Element)

				}
			case SearchRequest:
				d.Switchb(SearchDb)
				if req.Content == nil {
					response.Error = fmt.Errorf("No Search Parameter provided")
				} else {
					f, err := d.search(req.Content)
					if err != nil {
						response.Error = err
					} else {
						d.Switchb(DefaultDb)
						if test, err := d.c.Exists(f); err != nil || test <= 0 {
							if err != nil {
								response.Error = err
							} else {
								response.Error = fmt.Errorf("Data gone")
							}
						} else {
							elem, err := d.c.Get(f)
							if err != nil {
								response.Error = err
							} else {
								ctype := req.Content
								tools.Decode(elem, ctype)
								if ctype == nil {
									response.Error = fmt.Errorf("Data not Decoded")
								}
								response.Content = ctype
								response.Affected = f
							}

						}
					}

				}
			case ReadRequest:
				d.Switchb(DefaultDb)
				if req.Element == nil {
					response.Error = fmt.Errorf("No Element specified\n")
					//indexes := d.c.L
				} else {
					if test, err := d.c.Exists(req.Element); err != nil || test <= 0 {
						if err != nil {
							response.Error = err
						} else {
							response.Error = fmt.Errorf("Data not found to Read")
						}
					} else {
						da, err := d.c.Get(req.Element)
						if err != nil {
							response.Error = fmt.Errorf("ReadRequest Error from Ledis: %v for %v", err, req)
						} else {
							ctype := req.Content
							tools.Decode(da, ctype)
							if req.Content == nil {
								response.Error = fmt.Errorf("Data not Decoded")
							}
							response.Content = ctype
							response.Affected = req.Element
						}
					}
				}
			case DeleteRequest:
				d.c.Del(req.Element)
				response.Affected = req.Element
				d.Switchb(TypeDb)
				del := tools.ToBytes(DELETED)
				d.c.LPush(del, req.Element)
				d.Switchb(MetaDb)
				d.c.Del(req.Element)
				d.Switchb(DefaultDb)
			}
			con.ResponseChannel <- *response
			d.wg.Done()
			d.mtx.Unlock()
		}()
	}
	//return response
}

func (d *DataStorage) Switchb(db int) error {
	var e error
	d.c, e = d.l.Select(db)
	if e != nil {
		return e
	}
	return nil
}

func (d *DataStorage) Connect() *StorageConnection {
	c := new(StorageConnection)
	c.RequestChannel = make(chan StorageRequest)
	c.ResponseChannel = make(chan StorageResponse)
	d.clients.Add(1)
	d.clientCount++
	c.storage = d
	go func(c *StorageConnection) {
		d.Process(c.RequestChannel, c)
	}(c)
	return c
}

func (d *DataStorage) Serve() error {
	d.Running = true
	go func() {
		for {
			select {
			case cmd := <-d.controll:
				{
					if cmd == ShutdownCmd {
						d.Close()
						break
					}
				}
			}
		}
	}()
	return nil
}

func (d *DataStorage) Close() error {

	d.clients.Wait()
	var once sync.Once
	var onceBody = func() {
		d.Running = false
		d.wg.Wait()
		d.l.Close()

	}
	once.Do(onceBody)
	return nil
}

func OpenStorage(cfg *InitialConfiguration) (*DataStorage, error) {

	var ledcfg *lediscfg.Config
	if _, err := os.Stat(cfg.ConfigPath); os.IsNotExist(err) {
		ledcfg = lediscfg.NewConfigDefault()

		if len(cfg.DbPath) < 1 {
			//return nil, fmt.Errorf("Provide a DbPath in Configuration")
			panic("Provide a DbPath in Configuration")
		}

		if _, err := os.Stat(cfg.DbPath); os.IsNotExist(err) {
			err = os.MkdirAll(cfg.DbPath, cfg.PathAccessMode)
			if err != nil {
				return nil, err
			}
		}

		e := os.Chmod(cfg.DbPath, cfg.PathAccessMode)
		if e != nil {
			return nil, e
		}
		ledcfg.DataDir = cfg.DbPath
		if cfg.MaxDatabases > 0 {
			ledcfg.Databases = cfg.MaxDatabases
		}
		ledcfg.AuthPassword = cfg.AuthKey

		if _, err := os.Stat(cfg.ConfigPath); os.IsNotExist(err) {
			ledcfg.DumpFile(cfg.ConfigPath)
		}
	} else {
		ledcfg, err = lediscfg.NewConfigWithFile(cfg.ConfigPath)
		if err != nil {
			return nil, err
		}
	}

	if ledcfg != nil {
		var d = new(DataStorage)
		d.controll = make(chan int)
		var e error
		d.l, e = ledis.Open(ledcfg)
		if e != nil {
			panic(e)
		}

		d.c, e = d.l.Select(DefaultDb)
		if e != nil {
			panic(e)
		}

		d.clientCount = 0

		return d, e
	}
	return nil, fmt.Errorf("Failed to get Ledis Configuration")
}

func (d *DataStorage) Open(index int) error {
	t, err := d.l.Select(index)
	if err == nil {
		d.c = t
	}
	return err
}

func (d *DataStorage) Scope(scope string) {
	if d.transaction {
		panic("Can't change Scope while in Transaction")
	}
	d.scope = scope
}
