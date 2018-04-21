package tools

import (
	"strconv"
	"testing"
)

type TestData struct {
	Feld1 string
	Feld2 string
	Feld3 map[string]interface{}
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

func TestEncodeDecode(t *testing.T) {
	var sa8 = "嬨 鄨鎷闒 廅愮揫 嫀 趍跠跬 壾嵷幓 躨"
	var t1 string
	enc := EncodeToByte(sa8)
	Decode(enc, &t1)
	if sa8 != t1 {
		t.Errorf("Encoded and Decoded String doesn't compare %s != %s", sa8, t1)
	}

	var ui64 int = 1234353463643
	var t2 int
	enc = EncodeToByte(ui64)
	Decode(enc, &t2)
	if ui64 != t2 {
		t.Errorf("Encoded and Decoded Int doesn't compare")
	}

	var mui64 int = -1234353463643
	var t3 int
	enc = EncodeToByte(mui64)
	Decode(enc, &t3)
	if t3 != mui64 {
		t.Errorf("Encoded and Decoded Negative Int doesn't compare")
	}

	var bo bool = false
	var t4 bool
	enc = EncodeToByte(bo)
	Decode(enc, &t4)
	if bo != t4 {
		t.Errorf("Encoded and Decoded Bool doesn't compare")
	}

	var fl float64 = -245.3535345646453665565429
	var t5 float64
	enc = EncodeToByte(fl)
	Decode(enc, &t5)
	if fl != t5 {
		t.Errorf("Encoded and Decoded Float doesn't compare")
	}

	/*m := GetMap()
	var t6 map[string]interface{}
	enc = EncodeToByte(m)
	Decode(enc, &t6)

	if m["text"] != t6["text"] {
		t.Errorf("Maps do not Compare")
	}*/

	s := new(TestData)
	s.Feld1 = GetRandomAsciiString(10)
	s.Feld2 = GetRandomAsciiString(20)
	s.Feld3 = GetMap()
	enc = EncodeToByte(s)
	t7 := new(TestData)
	Decode(enc, &t7)

	if t7.Feld1 != s.Feld1 {
		t.Error("Decoded Field doned match to Orignal Data")
	}
}

func TestToBytes(t *testing.T) {

	var ui64 int = 1234353463643
	ui64b := ToBytes(ui64)
	nui64, _ := strconv.Atoi(string(ui64b))
	if nui64 != ui64 {
		t.Error("uint64 convert and back failed %d != %d", ui64, nui64)
	}

	var sa = "Gsdgrekgre oewkfgweo d"
	sab := ToBytes(sa)
	nsa := string(sab)
	if sa != nsa {
		t.Errorf("string convert and back failed %s != %s", sa, nsa)
	}

	var sa8 = "嬨 鄨鎷闒 廅愮揫 嫀 趍跠跬 壾嵷幓 躨"
	sa8b := ToBytes(sa8)
	nsa8 := string(sa8b[:])
	if sa8 != nsa8 {
		t.Errorf("string convert and back failed %s != %s", sa8, nsa8)
	}

	var bf bool = false
	bfb := ToBytes(bf)
	nbf, _ := strconv.ParseBool(string(bfb))
	if nbf != bf {
		t.Errorf("bool convert failed %b=%b", bf, nbf)
	}

}
