package routing_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/3devo/feconnector/models"
	"github.com/asdine/storm"
)

var logFiles = []models.LogFile{
	{
		UUID:      "550e8400-e29b-41d4-a716-446655440000",
		Name:      "log1",
		Timestamp: 1,
		HasNote:   false},
	{
		UUID:      "550e8400-e29b-41d4-a716-446655440001",
		Name:      "log2",
		Timestamp: 2,
		HasNote:   false},
	{
		UUID:      "550e8400-e29b-41d4-a716-446655440002",
		Name:      "log3",
		Timestamp: 3,
		HasNote:   true}}

var charts = []models.Chart{
	{
		UUID: "550e8400-e29b-41d4-a716-446655440000"},
	{
		UUID: "550e8400-e29b-41d4-a716-446655440001"},
	{
		UUID: "550e8400-e29b-41d4-a716-446655440002"}}

var sheets = []models.Sheet{
	{
		UUID:   "550e8400-e29b-41d4-a716-446655440000",
		Title:  "sheet0",
		Charts: []string{"550e8400-e29b-41d4-a716-446655440000"}},
	{
		UUID:  "550e8400-e29b-41d4-a716-446655440001",
		Title: "sheet1"},
	{
		UUID:  "550e8400-e29b-41d4-a716-446655440002",
		Title: "sheet2"}}

var workspaces = []models.Workspace{
	{
		UUID:   "550e8400-e29b-41d4-a716-446655440000",
		Title:  "workspace0",
		Sheets: []string{"550e8400-e29b-41d4-a716-446655440000"}},
	{
		UUID:  "550e8400-e29b-41d4-a716-446655440001",
		Title: "workspace1"},
	{
		UUID:  "550e8400-e29b-41d4-a716-446655440002",
		Title: "workspace2"}}

func PrepareDb() (string, *storm.DB) {
	dir := filepath.Join(os.TempDir(), "feconnector-test")
	os.MkdirAll(dir, os.ModePerm)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	os.Mkdir(filepath.Join(dir, "logs"), os.ModePerm)
	os.Mkdir(filepath.Join(dir, "notes"), os.ModePerm)
	for _, model := range logFiles {
		db.Save(&model)
	}
	for _, model := range logFiles {
		logName := model.Name + "-" + time.Unix(model.Timestamp, 0).Format("2006-01-02-15-04-05") + ".txt"
		f, _ := os.Create(filepath.Join(dir, "logs", logName))
		f.Close()
		if model.HasNote {
			ioutil.WriteFile(filepath.Join(dir, "notes", logName), []byte("test"), os.ModePerm)
		}
	}

	for i, _ := range charts {
		charts[i] = CreateChart(&charts[i])
		charts[i].Title = "chart" + strconv.Itoa(i)
		db.Save(&charts[i])
	}

	for i := range sheets {
		db.Save(&sheets[i])
	}

	for i := range workspaces {
		db.Save(&workspaces[i])
	}
	return dir, db
}

func CreateChart(chart *models.Chart) models.Chart {
	chart.Title = "test"
	chart.PlotDataInformation = []models.PlotDataInformation{
		{
			DataName: "dataName",
			PlotName: "plotName",
			Color:    "#4286f4",
			Axis:     ""}}
	chart.Axes = []models.Axis{
		{
			Name:  "xaxis",
			Title: "title",
			Range: []int{0},
		},
		{
			Name:  "yaxis",
			Title: "title",
			Range: []int{0, 2000}}}

	chart.HorizontalRulers = []models.Ruler{}
	chart.VerticalRulers = []models.Ruler{}
	chart.Image = "data:image/gif;base64,R0lG"
	return *chart
}
