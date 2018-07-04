package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/3devo/feconnector/models"
	"github.com/asdine/storm"
)

var devoCharts = []models.Chart{}

var devoSheets = []models.Sheet{}

var devoWorkspaces = []models.Workspace{}

func FillDatabase(Db *storm.DB) {
	defaultDir := "./default-files"
	charts, _ := ioutil.ReadDir(filepath.Join(defaultDir, "charts"))
	for _, file := range charts {
		chart, _ := ioutil.ReadFile(filepath.Join(defaultDir, "charts", file.Name()))
		temp := models.Chart{}
		json.Unmarshal(chart, &temp)
		Db.Save(&temp)
	}

	sheets, _ := ioutil.ReadDir(filepath.Join(defaultDir, "sheets"))
	for _, file := range sheets {
		sheet, _ := ioutil.ReadFile(filepath.Join(defaultDir, "sheets", file.Name()))
		temp := models.Sheet{}
		json.Unmarshal(sheet, &temp)
		Db.Save(&temp)
	}

	workspaces, _ := ioutil.ReadDir(filepath.Join(defaultDir, "workspaces"))
	for _, file := range workspaces {
		workspace, _ := ioutil.ReadFile(filepath.Join(defaultDir, "workspaces", file.Name()))
		temp := models.Workspace{}
		json.Unmarshal(workspace, &temp)
		Db.Save(&temp)
	}
}
