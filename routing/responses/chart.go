package responses

import "github.com/3devo/feconnector/models"

//swagger:parameters CreateChart UpdateChart
type ChartCreationParam struct {
	// in:body
	Data models.Chart `json:"data"`
}
