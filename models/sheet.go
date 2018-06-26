package models

type Sheet struct {
	UUID   string   `storm:"id" json:"uuid" validate:"uuid"`
	Title  string   `json:"title" validate:"required"`
	Charts []string `json:"charts" validate:"dive,chart-exists"`
}
