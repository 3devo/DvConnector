package responses

import (
	"github.com/3devo/dvconnector/models"
	"github.com/3devo/dvconnector/utils"
	"github.com/asdine/storm/q"
)

// WorkspaceResponse is a single logFile response model
//
// This is used for returning a response with a single sheet object as body
//
// swagger:model WorkspaceResponse
type WorkspaceResponse struct {
	UUID   string           `json:"uuid"`
	Title  string           `json:"title"`
	Sheets []*SheetResponse `json:"sheets"`
}

// WorkspaceCreationBody is a body that is needed to create workspaces through rest
// Parameters needed to create a sheet object
// swagger:parameters CreateWorkspace UpdateWorkspace
type WorkspaceCreationBody struct {
	//in:body
	Data models.Workspace
}

// GenerateWorkspaceResponseObject returns a WorkSpaceResponse object filled with actual sheet and chart data instead of id
func GenerateWorkspaceResponseObject(workspace *models.Workspace, env *utils.Env) *WorkspaceResponse {
	response := new(WorkspaceResponse)
	sheets := []models.Sheet{}
	selection := []q.Matcher{}

	for _, sheet := range workspace.Sheets {
		selection = append(selection, q.Eq("UUID", sheet))
	}

	query := env.Db.Select(q.Or(selection...))
	query.Find(&sheets)
	sheetResponses := make([]*SheetResponse, 0)

	for _, sheet := range sheets {
		sheetResponses = append(sheetResponses, GenerateSheetResponseObject(&sheet, env))
	}

	response.UUID = workspace.UUID
	response.Title = workspace.Title
	response.Sheets = sheetResponses
	return response
}
