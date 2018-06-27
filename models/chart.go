package models

type Axis struct {
	Name  string `json:"name" validate:"required"`
	Title string `json:"title" validate:"required"`
	Range []int  `json:"range" validate:"min=1"`
}

type PlotDataInformation struct {
	DataName string `json:"dataName" validate:"required"`
	PlotName string `json:"plotName" validate:"required"`
	Color    string `json:"color" validate:"required" validate:"hexcolor"`
	Axis     string `json:"axis" validate:"required,oneof y0 y1 y2"`
}

type Ruler struct {
	Text  string `json:"text" validate:"required"`
	Width int    `json:"width" validate:"required"`
	Color string `json:"color" validate:"hexcolor"`
}

// swagger:model Chart
type Chart struct {
	UUID                string                `storm:"id" json:"uuid" validate:"uuid"`
	Title               string                `json:"title" validate:"required"`
	PlotDataInformation []PlotDataInformation `json:"plotDataInformation"`
	Axes                []Axis                `json:"axes"`
	HorizontalRulers    []Ruler               `json:"horizontalRulers"`
	VerticalRulers      []Ruler               `json:"verticalRulers"`
	Image               string                `json:"image" validate:"base64"`
}
