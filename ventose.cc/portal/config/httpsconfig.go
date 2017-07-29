package config

import "ventose.cc/https"

type Mapping struct {
	Path string
	Address string
	OAuth bool
}

type PortalHttpsConfiguration struct {
	StaticFiles 	string
	Address 	string
	Ca		[]byte
	Cert		[]byte
	HasOAuth	bool
	Mappings	[]Mapping
}

func (c *PortalHttpsConfiguration) ToConfig() *https.HttpsConfiguration {
	cfg := &https.HttpsConfiguration{
		Address: 	c.Address,
		StaticPath: 	c.StaticFiles,
	}
	return cfg
}