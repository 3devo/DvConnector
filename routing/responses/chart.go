package responses

import "github.com/3devo/feconnector/models"

//swagger:parameters CreateChart UpdateChart
type ChartCreationParam struct {
	// in:body
	Body struct {
		UUID                string                       `json:"uuid"`
		Title               string                       `json:"title"`
		PlotDataInformation []models.PlotDataInformation `json:"plotDataInformation"`
		Axes                []models.Axis                `json:"axes"`
		HorizontalRulers    []models.Ruler               `json:"horizontalRulers"`
		VerticalRulers      []models.Ruler               `json:"verticalRulers"`
		Image               string                       `json:"image"`
	} `json:"body"`
}
