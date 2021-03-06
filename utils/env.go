package utils

import (
	"github.com/asdine/storm"
	validator "gopkg.in/go-playground/validator.v9"
)

type Env struct {
	DataDir   string
	ConfigDir string
	Db        *storm.DB
	Validator *validator.Validate
}
