package server

import (
	"testing"
	"ventose.cc/portal/config"
	"encoding/json"
	"io/ioutil"
	"ventose.cc/data"
	"ventose.cc/tools"
)

func createConfigT(t *testing.T) string {
	cfg := &config.PortalConfig{
		Contact: &config.PortalContact{
			Email: 		"osiegemund@gmail.com",
			Country: 	"World",
			City:		"Moon",
		},
		ConfigurationFile: "/tmp/test.cfg",
		Organisation: "Ventose Test",
		DB: &data.InitialConfiguration{
			ConfigPath: "/tmp/db.cfg",
			DbPath: "/tmp/db",
			MaxDatabases: 12,
			AuthKey: tools.GetRandomAsciiString(15),
			PathAccessMode: 0770,
		},
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ioutil.WriteFile(cfg.ConfigurationFile ,b, 0666)
	return cfg.ConfigurationFile
}

func createConfigB(t *testing.B) string {
	cfg := &config.PortalConfig{
		Contact: &config.PortalContact{
			Email: 		"osiegemund@gmail.com",
			Country: 	"World",
			City:		"Moon",
		},
		DB: &data.InitialConfiguration{
			ConfigPath: "/tmp/db.cfg",
			DbPath: "/tmp/db",
			MaxDatabases: 12,
			AuthKey: tools.GetRandomAsciiString(15),
			PathAccessMode: 0770,
		},
		ConfigurationFile: "/tmp/test.cfg",
		Organisation: "Ventose Test",
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ioutil.WriteFile(cfg.ConfigurationFile ,b, 0666)
	return cfg.ConfigurationFile
}


func BenchmarkStartPortal(b *testing.B) {
	cfgFile := createConfigB(b)
	for n := 0; n < b.N; n++ {
		_, err := StartPortal(cfgFile)
		if err != nil {
			b.Fatal(err)
		}
	}
}
func TestStartPortal(t *testing.T) {
	cfgFile := createConfigT(t)
	_, err := StartPortal(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
}
