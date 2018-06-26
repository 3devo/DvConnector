package models

type Workspace struct {
	UUID   string   `storm:"id" json:"uuid" validate:"uuid"`
	Title  string   `json:"title" validate:"required"`
	Sheets []string `json:"sheets" validate:"dive,sheet-exists"`
}
