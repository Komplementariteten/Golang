package tools

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/siddontang/go/bson"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	ENC_serialised = byte(1)
	ENC_string     = byte(2)
	ENC_float      = byte(3)
	ENC_int        = byte(4)
	ENC_byteslice  = byte(5)
	ENC_bool       = byte(6)
	letterBytes    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	OSSP           = string(os.PathSeparator)
)

var Base64Encode = base64.RawURLEncoding

func GetRandomAsciiString(size int) string {
	bytes := GetRandomBytes(size)
	for i, b := range bytes {
		bytes[i] = letterBytes[b%byte(len(letterBytes))]
	}
	return string(bytes)
}

func GetRandomBytes(size int) []byte {
	c := make([]byte, size)
	rand.Read(c)
	return c
}

func TestStringInput(str string) string {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		readen := scanner.Text()
		if strings.ContainsAny(str, readen) {
			return readen
		}
	}
	return ""
}

func GetDirSeperator() string {
	return string(os.PathSeparator)
}

func GetSerial() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)
	return serialNumber
}

func Itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func Decode(b []byte, v interface{}) {
	typeFlag := b[0]
	data := b[1:]
	var e error
	switch typeFlag {
	default:
		panic(fmt.Errorf("Type unkown, can't unserialise %t", typeFlag))
	case ENC_serialised:
		e = bson.Unmarshal(data, v)
		if e != nil {
			panic(fmt.Errorf("Failed to Unmarshal %v with %s", v, e.Error()))
		}
		//enc, err := json.Marshal()
	case ENC_string:
		s := v.(*string)
		*s = string(data)
	case ENC_int:
		s := v.(*int)
		*s, e = strconv.Atoi(string(data))
		if e != nil {
			panic(e)
		}
	case ENC_bool:
		s := v.(*bool)
		*s, e = strconv.ParseBool(string(data))
		if e != nil {
			panic(e)
		}
	case ENC_float:
		s := v.(*float64)
		i := binary.LittleEndian.Uint64(data)
		*s = math.Float64frombits(i)
	}
}

func EncodeToBytes(t interface{}) []byte {
	a := ToBytes(t)
	n := len(a) + 1
	b := make([]byte, n, n)
	switch t.(type) {
	default:
		b[0] = ENC_serialised
		copy(b[1:n], a)
	case string:
		b[0] = ENC_string
		copy(b[1:n], a)
	case int:
		b[0] = ENC_int
		copy(b[1:n], a)
	case float64, float32:
		b[0] = ENC_float
		copy(b[1:n], a)
	case []byte:
		b[0] = ENC_byteslice
		copy(b[1:n], a)
	case bool:
		b[0] = ENC_bool
		copy(b[1:n], a)
	}
	return b
}
func ToBytes(t interface{}) []byte {

	var buff = new(bytes.Buffer)

	switch t.(type) {
	default:
		bd, err := bson.Marshal(t)
		if err != nil {
			panic(fmt.Errorf("Failed to Marshal %v with %s", t, err.Error()))
		}
		return bd
	case string:
		return []byte(t.(string))
	case int, uint:
		istr := strconv.Itoa(t.(int))
		return []byte(istr)
	case float64:
		fl := math.Float64bits(t.(float64))
		e := binary.Write(buff, binary.LittleEndian, fl)
		if e != nil {
			log.Fatal(e)
		}
		return buff.Bytes()
	case []byte:
		return t.([]byte)
	case bool:
		if t.(bool) {
			return []byte("true")
		} else {
			return []byte("false")
		}
	}
	return nil
}

func LoadFromJsonFile(path string, t interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Failed to read JSON File %s with: %v", path, err)
	}
	err = json.Unmarshal(data, t)
	if err != nil {
		return fmt.Errorf("Failed to Parse JSON from %s with: %v", path, err)
	}
	return nil
}

func DumpConfig(path string, t interface{}) error {
	jbytes, err := json.Marshal(t)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, jbytes, 0640)
	return err
}

func StartWorker(f func(a ...interface{}), a ...interface{}) error {
	go func(a ...interface{}) {
		f(a)
	}(a)
	return nil
}

func IsSlice(arg interface{}) (reflect.Value, bool) {
	val := reflect.ValueOf(arg)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		return val, true
	}
	return val, false
}
