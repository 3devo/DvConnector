package models

import "github.com/3devo/feconnector/utils"

type Config struct {
	ID                int    `storm:"id" json:"-"`
	SpectatorsAllowed bool   `json:"spectatorsAllowed"`
	AuthRequired      bool   `json:"authRequired"`
	Password          []byte `json:"password,omitempty"`
}

func (config *Config) CheckPasswordhash(password string) bool {
	return utils.ComparePasswords(config.Password, []byte(password))
}

func NewConfig(authRequired bool, spectatorsAllowed bool, password string) *Config {
	config := new(Config)
	config.ID = 1
	config.SpectatorsAllowed = spectatorsAllowed
	config.AuthRequired = authRequired
	config.Password = utils.HashAndSalt(password)
	return config
}
