package models

import (
	"github.com/3devo/feconnector/utils"
	"github.com/asdine/storm/q"
)

type Workspace struct {
	ID     int    `storm:"id,increment"`
	Title  string `json:"title"`
	Sheets []int  `json:"sheets"`
}

type WorkspaceResponse struct {
	ID     int
	Title  string `json:"title"`
	Sheets []SheetResponse
}

func (workspace *Workspace) GetResponseObject(env *utils.Env) WorkspaceResponse {
	var sheets []Sheet
	var selection []q.Matcher
	for _, sheet := range workspace.Sheets {
		selection = append(selection, q.Eq("ID", sheet))
	}
	query := env.Db.Select(q.Or(selection...))
	query.Find(&sheets)
	sheetResponses := make([]SheetResponse, 0)
	for _, sheet := range sheets {
		sheetResponses = append(sheetResponses, sheet.GetResponseObject(env))
	}

	return WorkspaceResponse{
		ID:     workspace.ID,
		Title:  workspace.Title,
		Sheets: sheetResponses}
}
