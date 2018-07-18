package models

type BlackListedToken struct {
	ID         int    `storm:"id,increment"`
	Token      string `storm:"unique"`
	Expiration int64
}
