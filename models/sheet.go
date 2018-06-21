package models

import (
	"github.com/3devo/feconnector/utils"
	"github.com/asdine/storm/q"
)

type Sheet struct {
	ID     int    `storm:"id,increment"`
	Title  string `json:"title"`
	Charts []int  `json:"charts"`
}

type SheetResponse struct {
	ID     int
	Title  string `json:"title"`
	Charts []Chart
}

func (sheet *Sheet) GetResponseObject(env *utils.Env) SheetResponse {
	var charts []Chart
	var selection []q.Matcher
	for _, chart := range sheet.Charts {
		selection = append(selection, q.Eq("ID", chart))
	}
	query := env.Db.Select(q.Or(selection...))
	query.Find(&charts)

	return SheetResponse{
		ID:     sheet.ID,
		Title:  sheet.Title,
		Charts: charts}
}
