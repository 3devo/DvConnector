package responses

import (
	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/utils"
	"github.com/asdine/storm/q"
)

// A single logFile response model
//
// This is used for returning a response with a single sheet object as body
//
// swagger:model SheetResponse
type SheetResponse struct {
	UUID   string         `json:"uuid"`
	Title  string         `json:"title"`
	Charts []models.Chart `json:"charts"`
}

// Parameters needed to create a sheet object
// swagger:parameters CreateSheet UpdateSheet
type SheetCreationParams struct {
	//in:body
	Data models.Sheet `json:"data"`
}

// GenerateSheetResponseObject returns a SheetResponse object filled with actual chart data instead of id
func GenerateSheetResponseObject(sheet *models.Sheet, env *utils.Env) *SheetResponse {
	var charts []models.Chart
	var selection []q.Matcher
	response := new(SheetResponse)

	for _, chart := range sheet.Charts {
		selection = append(selection, q.Eq("UUID", chart))
	}
	query := env.Db.Select(q.Or(selection...))
	query.Find(&charts)

	response.UUID = sheet.UUID
	response.Title = sheet.Title
	response.Charts = charts
	return response
}
