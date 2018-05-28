package utils

import (
	"log"

	"github.com/asdine/storm"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func HashAndSalt(pwd string) []byte {

	// Use GenerateFromPassword to hash & salt pwd
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return hash
}

func ComparePasswords(hashedPwd []byte, plainPwd []byte) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	log.Print(hashedPwd)
	log.Print(plainPwd)
	err := bcrypt.CompareHashAndPassword(hashedPwd, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

type Env struct {
	Db           *storm.DB
	SessionStore *sessions.CookieStore
}
