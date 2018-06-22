package models

type Workspace struct {
	UUID   string   `storm:"id" json:"uuid"`
	Title  string   `json:"title"`
	Sheets []string `json:"sheets"`
}
