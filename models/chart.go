package models

type Axis struct {
	Name       string `json:"name" validate:"required"`
	Title      string `json:"title" validate:"required"`
	Range      []int  `json:"range" validate:"min=1"`
	Side       string `json:"side"`
	Overlaying string `json:"overlaying" validate:"oneof y x"`
}

type PlotDataInformation struct {
	DataName string `json:"dataName" validate:"required"`
	PlotName string `json:"plotName" validate:"required"`
	Color    string `json:"color" validate:"required" validate:"hexcolor"`
	Axis     string `json:"axis" validate:"oneof y0 y1 y2"`
}

type Ruler struct {
	Text   string `json:"text" validate:"required"`
	Width  int    `json:"width" validate:"required"`
	Color  string `json:"color" validate:"hexcolor"`
	Filler bool   `json:"filler"`
}
type HRuler struct {
	Ruler
	Y int `json:"y" validate:"required"`
}

type VRuler struct {
	Ruler
	X int `json:"x" validate:"required"`
}

// Chart with the needed properties to generate a chart in the frontend
// swagger:model Chart
type Chart struct {
	UUID                string                `storm:"id" json:"uuid" validate:"uuid"`
	Title               string                `json:"title" validate:"required"`
	PlotDataInformation []PlotDataInformation `json:"plotDataInformation"`
	Axes                []Axis                `json:"axes"`
	HorizontalRulers    []HRuler              `json:"horizontalRulers"`
	VerticalRulers      []VRuler              `json:"verticalRulers"`
	Image               string                `json:"image" validate:"datauri"`
}
