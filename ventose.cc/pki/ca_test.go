package pki

import "os"
import (
	"ventose.cc/data"
)

const (
	DbCfg  = "/tmp/ledif2.cfg"
	DbPath = "/tmp/testdb2"
)

func GetTestCfg() *data.InitialConfiguration {
	cfg := new(data.InitialConfiguration)
	cfg.ConfigPath = DbCfg
	cfg.AuthKey = "12345"
	cfg.DbPath = DbPath
	cfg.MaxDatabases = 10
	cfg.PathAccessMode = os.FileMode(0700)
	return cfg
}
