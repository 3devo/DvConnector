package models

type Sheet struct {
	UUID   string   `storm:"id" json:"uuid"`
	Title  string   `json:"title"`
	Charts []string `json:"charts"`
}
