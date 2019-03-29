package responses

import "github.com/3devo/dvconnector/models"

// ChartCreationBody is the body needed to create a chart through rest
// swagger:parameters CreateChart UpdateChart
type ChartCreationBody struct {
	// in:body
	Data models.Chart `json:"data"`
}
