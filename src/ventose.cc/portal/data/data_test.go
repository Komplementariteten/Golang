package data

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"ventose.cc/tools"
)

type TestType struct {
	a int
	b string
	c float64
}

type TestData struct {
	RequestContent
	Operation   int
	Name        string `search:"name"`
	Id          int
	Description string
}

func GetMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["abc"] = 2.2345124
	m["ttse"] = 234.234524
	m["arr"] = []int{1, 2, 3, 4}
	m["text"] = "嬨 鄨鎷闒 廅愮揫 嫀 趍跠跬 壾嵷幓 躨"
	//m["type"] = &TestType{1, "abc", 0.0000001}
	m["id"] = 123
	return m
}

const (
	DbCfg  = "/tmp/ledif.cfg"
	DbPath = "/tmp/testdb"
)

func GetTestCfg() *InitialConfiguration {
	cfg := new(InitialConfiguration)
	cfg.ConfigPath = DbCfg
	cfg.AuthKey = "12345"
	cfg.DbPath = DbPath
	cfg.MaxDatabases = 10
	cfg.PathAccessMode = os.FileMode(0700)
	return cfg
}

func BenchmarkConnect(b *testing.B) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer d.Close()
	d.Serve()
	con := d.Connect()
	for i := 0; i < b.N; i++ {
		r := new(StorageRequest)
		con.RequestChannel <- *r
		<-con.ResponseChannel
	}
	CleanTestDb()
}

func BenchmarkDel(b *testing.B) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer d.Close()
	d.Serve()
	con := d.Connect()
	// Setup Data
	da := new(TestData)
	da.Description = "ABC"
	da.Name = "DEF"
	da.Id = 12
	for i := 0; i < b.N; i++ {
		req := new(StorageRequest)
		req.Type = CreateRequest
		req.Content = da
		con.RequestChannel <- *req
		resp := <-con.ResponseChannel

		req.Type = DeleteRequest
		req.Element = resp.Affected
		con.RequestChannel <- *req
		<-con.ResponseChannel
		if resp.Error != nil {
			b.Fatal(resp.Error)
		}
	}
	CleanTestDb()
}

func BenchmarkAdd(b *testing.B) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer d.Close()
	d.Serve()
	con := d.Connect()
	// Setup Data
	da := new(TestData)
	da.Description = "ABC"
	da.Name = "DEF"
	da.Id = 12
	for i := 0; i < b.N; i++ {
		req := new(StorageRequest)
		req.Type = CreateRequest
		req.Content = da
		con.RequestChannel <- *req
		resp := <-con.ResponseChannel

		if resp.Error != nil {
			b.Fatal(resp.Error)
		}
		fmt.Printf("ID: %v\n", resp.Affected)
	}
	CleanTestDb()
}

func BenchmarkUpdate(b *testing.B) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer d.Close()
	d.Serve()
	con := d.Connect()
	// Setup Data
	da := new(TestData)
	da.Description = "ABC"
	da.Name = "DEF"
	da.Id = 12
	for i := 0; i < b.N; i++ {
		req := new(StorageRequest)
		req.Type = CreateRequest
		req.Content = da
		con.RequestChannel <- *req
		resp := <-con.ResponseChannel

		if resp.Error != nil {
			b.Fatal(resp.Error)
		}

		req.Element = resp.Affected
		req.Type = UpdateRequest
		req.Content = da
		da.Description = tools.GetRandomAsciiString(12)
		con.RequestChannel <- *req
		uresp := <-con.ResponseChannel

		if uresp.Error != nil {
			b.Fatal(resp.Error)
		}

		if !bytes.Equal(uresp.Affected, resp.Affected) {
			b.Errorf("Wrong Element Updated %x != %x", uresp.Affected, resp.Affected)
		}
	}
	CleanTestDb()
}

func BenchmarkList(b *testing.B) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer d.Close()
	d.Serve()
	con := d.Connect()

	var max int64
	max = 10

	for i := 0; i < int(max); i++ {
		da := new(TestData)
		da.Description = tools.GetRandomAsciiString(6)
		da.Name = tools.GetRandomAsciiString(7)
		da.Id = 12

		// Setup Request
		req := new(StorageRequest)
		req.Type = CreateRequest
		req.Content = da
		con.RequestChannel <- *req
		resp := <-con.ResponseChannel
		if resp.Error != nil {
			b.Fatal(resp.Error)
		}

	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := new(StorageRequest)
		req.Type = ListRequest
		req.Parameter = new(ParameterType)
		req.Parameter.Offset = 1
		req.Parameter.Top = 14
		req.Content = new(TestData)
		con.RequestChannel <- *req
		resp := <-con.ResponseChannel
		if resp.Error != nil {
			b.Fatal(resp.Error)
		}
	}
	CleanTestDb()
}

func TestSearch(t *testing.T) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	defer d.Close()
	if err != nil {
		t.Fatal(err)
	}
	d.Serve()

	con := d.Connect()
	defer con.Close()
	da := new(TestData)
	da.Description = tools.GetRandomAsciiString(6)
	da.Name = tools.GetRandomAsciiString(12)
	da.Id = 71

	req := new(StorageRequest)
	req.Type = CreateRequest
	req.Content = da
	con.RequestChannel <- *req
	resp := <-con.ResponseChannel
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	s := new(TestData)
	s.Name = da.Name

	req = new(StorageRequest)
	req.Type = SearchRequest
	req.Content = da

	con.RequestChannel <- *req
	resp = <-con.ResponseChannel

	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	found, ok := resp.Content.(*TestData)
	if !ok {
		t.Fatal("Can't Assert Data %v:%v", found, resp.Content)
	}

	if found.Name != da.Name {
		t.Fatal("Found wrong data")
	}
}

func TestDescribe(t *testing.T) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	defer d.Close()
	if err != nil {
		t.Fatal(err)
	}
	d.Serve()

	con := d.Connect()
	defer con.Close()
	da := new(TestData)
	da.Description = tools.GetRandomAsciiString(6)
	da.Name = tools.GetRandomAsciiString(7)
	da.Id = 71

	req := new(StorageRequest)
	req.Type = CreateRequest
	req.Content = da
	con.RequestChannel <- *req
	resp := <-con.ResponseChannel
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	req = new(StorageRequest)
	req.Type = DescribeRequest
	req.Element = resp.Affected
	con.RequestChannel <- *req
	resp2 := <-con.ResponseChannel
	if resp2.Error != nil {
		t.Fatal(resp2.Error)
	}

	if meta, ok := resp2.Content.(*DescribeResponse); ok {
		if !meta.Found {
			t.Fatal("Newly added Item can't be described")
		}
	} else {

		t.Fatalf("Describe Request does not Respond with DesciribeResponse Type: %v", resp2)
	}

	req = new(StorageRequest)
	req.Type = CreateRequest
	req.Content = da
	con.RequestChannel <- *req
	resp = <-con.ResponseChannel
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	req = new(StorageRequest)
	req.Type = DescribeRequest
	req.Content = new(TestData)
	req.Element = nil
	con.RequestChannel <- *req
	resp3 := <-con.ResponseChannel
	if resp3.Error != nil {
		t.Fatal(resp3.Error)
	}

	if meta, ok := resp3.Content.(*DescribeResponse); ok {
		if meta.Count != 3 {
			t.Fatalf("Newly added Item not Counted. Count is %v", meta.Count)
		}
	} else {
		t.Fatal("Describe Request does not Respond with DesciribeResponse Type")
	}

	CleanTestDb()
}

func TestPersistence(t *testing.T) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	d.Serve()

	con := d.Connect()

	defaultId := 64

	da := new(TestData)
	da.Description = tools.GetRandomAsciiString(6)
	da.Name = tools.GetRandomAsciiString(7)
	da.Id = defaultId

	req := new(StorageRequest)
	req.Type = CreateRequest
	req.Content = da
	con.RequestChannel <- *req
	resp := <-con.ResponseChannel
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	con.Close()
	d.Close()

	d, err = OpenStorage(cfg)
	defer d.Close()
	if err != nil {
		t.Fatal(err)
	}
	d.Serve()

	con = d.Connect()
	defer con.Close()

	req = new(StorageRequest)
	req.Type = ReadRequest

	reqData := new(TestData)
	req.Content = reqData
	req.Element = resp.Affected

	con.RequestChannel <- *req
	resp2 := <-con.ResponseChannel

	if !bytes.Equal(resp.Affected, resp2.Affected) {
		t.Fatal("Created Element is not Readen %v!=%v", resp.Affected, resp2.Affected)
	}

	dat := req.Content.(*TestData)

	if dat.Id != defaultId {
		t.Fatal("Readen Data don't contain correct Content")
	}
	CleanTestDb()
}

func TestConcurrency(t *testing.T) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	d.Serve()

	var max int
	max = 10

	for i := 0; i < max; i++ {
		con := d.Connect()
		defer con.Close()

		go func(con *StorageConnection) {

			da := new(TestData)
			da.Description = tools.GetRandomAsciiString(6)
			da.Name = tools.GetRandomAsciiString(7)
			da.Id = 11

			req := new(StorageRequest)
			req.Type = CreateRequest
			req.Content = da
			con.RequestChannel <- *req
			resp := <-con.ResponseChannel
			if resp.Error != nil {
				t.Fatal(resp.Error)
			}

			da.Name = tools.GetRandomAsciiString(12)
			req.Type = UpdateRequest
			req.Content = da
			req.Element = resp.Affected
			con.RequestChannel <- *req

			resp = <-con.ResponseChannel

			if resp.Error != nil {
				t.Fatal(resp.Error)
			}

			req.Type = ReadRequest

			con.RequestChannel <- *req

			resp = <-con.ResponseChannel

			if resp.Error != nil {
				t.Fatal(resp.Error)
			}

			req.Type = DeleteRequest
			con.RequestChannel <- *req
			resp = <-con.ResponseChannel

		}(con)

	}
	CleanTestDb()

}

func TestList(t *testing.T) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	d.Serve()
	con := d.Connect()
	defer con.Close()

	var max int64
	max = 10

	var firstName string

	for i := 0; i < int(max); i++ {
		da := new(TestData)
		da.Description = tools.GetRandomAsciiString(6)
		da.Name = tools.GetRandomAsciiString(7)
		da.Id = 12

		if i == 0 {
			firstName = da.Name
		}

		// Setup Request
		req := new(StorageRequest)
		req.Type = CreateRequest
		req.Content = da
		con.RequestChannel <- *req
		resp := <-con.ResponseChannel
		if resp.Error != nil {
			t.Fatal(resp.Error)
		}
	}

	req := new(StorageRequest)
	req.Type = ListRequest
	req.Parameter = new(ParameterType)
	req.Parameter.Offset = 1
	req.Parameter.Top = 4
	req.Content = new(TestData)
	con.RequestChannel <- *req
	resp := <-con.ResponseChannel
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}
	listResp := resp.Content.(*ListResponse)
	if listResp.Count != max {
		t.Fatal("List Size does not Match ListResponse Size %d != max", listResp.Count, max)
	}

	for _, el := range listResp.Content {
		item := el.(*TestData)
		if item.Id != 12 {
			t.Fatal("Readen Item Content test failed")
		}
	}

	first := listResp.Content[0]
	firstItem := first.(*TestData)

	if firstName == firstItem.Name {
		t.Fatal("First Item returned is first Inserted, but should be secound!")
	}

	ln := len(listResp.Content)
	if ln != 4 {
		t.Fatal("Listed Items count does not match Requested")
	}
	CleanTestDb()
}

func TestRead(t *testing.T) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	d.Serve()

	con := d.Connect()
	defer con.Close()

	// Setup Data
	da := new(TestData)
	da.Description = "ABC"
	da.Name = "DEF"
	da.Id = 12

	// Setup Request
	req := new(StorageRequest)
	req.Type = CreateRequest
	req.Content = da
	con.RequestChannel <- *req
	resp := <-con.ResponseChannel
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	rreq := new(StorageRequest)
	rreq.Type = ReadRequest
	rreq.Element = resp.Affected
	rreq.Content = new(TestData)
	con.RequestChannel <- *rreq

	rresp := <-con.ResponseChannel

	if rresp.Error != nil {
		t.Fatal(rresp.Error)
	}

	if rresp.Content == nil {
		t.Fatal("no Data Returned")
	}

	if len(rresp.Affected) != INDEX_SIZE {
		t.Fatal("Can't ident Data, Wrong key Size")
	}

	data := rresp.Content.(*TestData)
	if data.Id != da.Id {
		t.Fatal("Readen Data dosen't match Written")
	}
	CleanTestDb()
}

func TestUpdate(t *testing.T) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	d.Serve()
	con := d.Connect()
	defer con.Close()

	// Setup Data
	da := new(TestData)
	da.Description = "ABC"
	da.Name = "DEF"
	da.Id = 12

	// Setup Request
	req := new(StorageRequest)
	req.Type = CreateRequest
	req.Content = da
	con.RequestChannel <- *req
	resp := <-con.ResponseChannel
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	if len(resp.Affected) != INDEX_SIZE {
		t.Fatal("Add didn't return correct index size")
	}

	req.Type = UpdateRequest
	req.Element = resp.Affected
	da.Name = "hhhgfskkldfkl"
	req.Content = da
	con.RequestChannel <- *req
	dresp := <-con.ResponseChannel

	if dresp.Error != nil {
		t.Fatal(dresp.Error)
	}

	if !bytes.Equal(resp.Affected, dresp.Affected) {
		t.Error("Delet Returned wrong index")
	}

	CleanTestDb()
}

func TestDel(t *testing.T) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	d.Serve()
	con := d.Connect()
	defer con.Close()
	// Setup Data
	da := new(TestData)
	da.Description = "ABC"
	da.Name = "DEF"
	da.Id = 12

	// Setup Request
	req := new(StorageRequest)
	req.Type = CreateRequest
	req.Content = da
	con.RequestChannel <- *req
	resp := <-con.ResponseChannel
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}
	if len(resp.Affected) != INDEX_SIZE {
		t.Fatal("Add didn't return correct index size")
	}

	req.Type = DeleteRequest
	req.Element = resp.Affected
	con.RequestChannel <- *req
	dresp := <-con.ResponseChannel

	if !bytes.Equal(resp.Affected, dresp.Affected) {
		t.Error("Delet Returned wrong index")
	}

	CleanTestDb()
}

func TestAdd(t *testing.T) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	d.Serve()

	con := d.Connect()
	defer con.Close()

	// Setup Data
	da := new(TestData)
	da.Description = "ABC"
	da.Name = "DEF"
	da.Id = 12

	// Setup Request
	req := new(StorageRequest)
	req.Type = CreateRequest
	req.Content = da
	con.RequestChannel <- *req
	resp := <-con.ResponseChannel
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}
	if len(resp.Affected) != INDEX_SIZE {
		t.Fatal("Add diden't return correct index size")
	}
	CleanTestDb()
}

func TestConnect(t *testing.T) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	defer d.Close()
	if err != nil {
		t.Fatal(err)
	}
	d.Serve()
	con := d.Connect()
	defer con.Close()

	r := new(StorageRequest)
	con.RequestChannel <- *r
	<-con.ResponseChannel
	CleanTestDb()
}

func TestBackground(t *testing.T) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	defer d.Close()
	if err != nil {
		t.Fatal(err)
	}
	d.Serve()
	if !d.Running {
		t.Errorf("Can't determin Server is running")
	}
	for i := 0; i < 100; i++ {

	}
	CleanTestDb()

}

func TestStartup(t *testing.T) {
	cfg := GetTestCfg()
	d, err := OpenStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	d.Close()

	b, err := OpenStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	CleanTestDb()
}

func CleanTestDb() {
	os.RemoveAll(DbCfg)
	os.RemoveAll(DbPath)
}

/*
func TestPlainLedis(t *testing.T) {
	cfg := lediscfg.NewConfigDefault()
	cfg.DataDir = "/tmp/ledis-test.db"
	cfg.Databases = 1024
	os.RemoveAll("/tmp/ledis-test.db")

	con, err := ledis.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer con.Close()
	c, err := con.Select(0)
	if err != nil{
		t.Fatal(err)
	}

	//Test Key Value
	err = c.Set(tools.ToBytes('Ä'), tools.ToBytes("äöpü"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Get(tools.ToBytes('Ä'))
	if err != nil {
		t.Fatal(err)
	}
	c.Del(tools.ToBytes('Ä'))
	index := tools.GetRandomBytes(25)

	if len(index) != 25 {
		t.Fatal("GetRandomBytes Returns wrong sizes")
	}
	m := GetMap()
	b := tools.EncodeToByte(m)

	err = c.Set(index, b)

	v, err := c.Get(index)
	if err != nil {
		t.Fatal(err)
	}
	var t1 map[string]interface{}
	tools.Decode(v, &t1)

	if t1["text"] != m["text"] {
		t.Errorf("Maps doen't compare after reading from db")
	}

	c.Del(index)


}*/

/*func TestLedis(t *testing.T) {

	m := GetMap()
	db, err := Load("testing.db")
	if err != nil {
		t.Error(err)
	}
	db.Open(0)
	db.Scope("t")

	for k, v := range m {
		db.c.
	}

}*/

/*func TestMap(t *testing.T) {
	db, err := Load("testing.db")
	if err != nil {
		t.Error(err)
	}
	db.Open(0)
	db.Scope("test")
	merr := db.AddMap(m)
	if merr != nil {
		t.Error(merr)
	}

	db.GetByValue("id", 123)
}*/
