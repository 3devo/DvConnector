package models

type User struct {
	UUID            string `storm:"id" json:"uuid" validate:"uuid"`
	Username        string `json:"username" storm:"unique" validate:"email"`
	Password        string `json:"password" validate:"min=8"`
	TrackingAllowed bool   `json:"trackingAllowed"`
}
