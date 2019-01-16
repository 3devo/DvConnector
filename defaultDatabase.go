package main

import (
	"encoding/json"
	"strings"

	"github.com/3devo/feconnector/models"
	"github.com/asdine/storm"
	packr "github.com/gobuffalo/packr/v2"
)

// FillDatabase fills the database with default values
func FillDatabase(Db *storm.DB) {
	box := packr.New("default-files", "./default-files")
	temp := models.Config{
		OpenNetwork: false}
	Db.Save(&temp)
	for _, file := range box.List() {
		fileBody, _ := box.Find(file)
		if strings.HasPrefix(file, "charts") {
			temp := models.Chart{}
			json.Unmarshal(fileBody, &temp)
			Db.Save(&temp)

		}
		if strings.HasPrefix(file, "sheets") {
			temp := models.Sheet{}
			json.Unmarshal(fileBody, &temp)
			Db.Save(&temp)

		}
		if strings.HasPrefix(file, "workspaces") {
			temp := models.Workspace{}
			json.Unmarshal(fileBody, &temp)
			Db.Save(&temp)
		}
	}
}
