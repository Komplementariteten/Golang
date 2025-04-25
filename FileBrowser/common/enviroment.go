package common

import "os"

type EnvSingelton struct {
	BaseDir string
}

var env *EnvSingelton

func Enviroment() *EnvSingelton {
	if env == nil {
		cwd, _ := os.Getwd()
		env = &EnvSingelton{
			BaseDir: cwd,
		}
	}
	return env
}

func (*EnvSingelton) SetWd(path string) {
	env.BaseDir = path
}
