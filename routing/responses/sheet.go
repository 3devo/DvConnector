package responses

import (
	"github.com/3devo/dvconnector/models"
	"github.com/3devo/dvconnector/utils"
	"github.com/asdine/storm/q"
)

// SheetResponse is a single logFile response model
//
// This is used for returning a response with a single sheet object as body
//
// swagger:model SheetResponse
type SheetResponse struct {
	UUID   string         `json:"uuid"`
	Title  string         `json:"title"`
	Charts []models.Chart `json:"charts"`
}

// SheetCreationBody is the body that is needed to create a new sheet through rest
// swagger:parameters CreateSheet UpdateSheet
type SheetCreationBody struct {
	//in:body
	Data models.Sheet `json:"data"`
}

// GenerateSheetResponseObject returns a SheetResponse object filled with actual chart data instead of id
func GenerateSheetResponseObject(sheet *models.Sheet, env *utils.Env) *SheetResponse {
	charts := []models.Chart{}
	selection := []q.Matcher{}
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
