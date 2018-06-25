package models

type Axis struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Range []int  `json:"range"`
}

type PlotDataInformation struct {
	DataName string `json:"dataName"`
	PlotName string `json:"plotName"`
	Color    string `json:"color"`
	Axis     string `json:"axis"`
}

type Ruler struct {
	Text  string `json:"text"`
	Width int    `json:"width"`
	Color string `json:"color"`
}

// swagger:model Chart
type Chart struct {
	UUID                string                `storm:"id" json:"uuid" validate:"uuid"`
	Title               string                `json:"title" validate:"required"`
	PlotDataInformation []PlotDataInformation `json:"plotDataInformation"`
	Axes                []Axis                `json:"axes"`
	HorizontalRulers    []Ruler               `json:"horizontalRulers"`
	VerticalRulers      []Ruler               `json:"verticalRulers"`
	Image               string                `json:"image"`
}
