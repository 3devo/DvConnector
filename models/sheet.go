package models

import (
	"github.com/3devo/feconnector/utils"
	"github.com/asdine/storm/q"
)

type Sheet struct {
	UUID   string `storm:"id" json:"uuid"`
	Title  string `json:"title"`
	Charts []int  `json:"charts"`
}

type SheetResponse struct {
	UUID   string `storm:"id" json:"uuid"`
	Title  string `json:"title"`
	Charts []Chart
}

func (sheet *Sheet) GetResponseObject(env *utils.Env) SheetResponse {
	var charts []Chart
	var selection []q.Matcher
	for _, chart := range sheet.Charts {
		selection = append(selection, q.Eq("UUID", chart))
	}
	query := env.Db.Select(q.Or(selection...))
	query.Find(&charts)

	return SheetResponse{
		UUID:   sheet.UUID,
		Title:  sheet.Title,
		Charts: charts}
}
