package config

import "ventose.cc/data"

type PortalContact struct {
	City    string
	County  string
	Country string
	Email   string
	Address string
}

type PortalConfig struct {
	Organisation      string
	Contact           *PortalContact
	PKI               *PkiConfiguration
	ConfigurationFile string
	DB                *data.InitialConfiguration
}
