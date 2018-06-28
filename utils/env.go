package utils

import (
	"github.com/asdine/storm"
	"gopkg.in/go-playground/validator.v9"
)

type Env struct {
	FileDir   string
	Db        *storm.DB
	Validator *validator.Validate
}
