package responses

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/utils"
)

// LogFileResponse is a single logFile response model
//
// This is used for returning a response with a single order as body
//
// swagger:response LogFileResponse
type LogFileResponse struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
	Note      string `json:"note"`
	Log       string `json:"log"`
}

// LogFileCreationBody is a model for creating logfiles through rest
// This is used to validate the update request
// swagger:parameters UpdateLogFile CreateLogFile
type LogFileCreationBody struct {
	// in:body
	Data struct {
		// required
		UUID string `json:"uuid" validate:"uuid"`
		Name string `json:"name" validate:"required"`
		Note string `json:"note"`
	} `json:"data"`
}

// GenerateLogResponse returns a new LogFileResponse filled with the actual note and log data
func GenerateLogResponse(logFile *models.LogFile, env *utils.Env) *LogFileResponse {
	response := new(LogFileResponse)
	response.Name = logFile.Name
	response.Timestamp = logFile.Timestamp
	response.UUID = logFile.UUID
	logName := logFile.Name + "-" + time.Unix(logFile.Timestamp, 0).Format("2006-01-02-15-04-05") + ".txt"
	logData, err := ioutil.ReadFile(fmt.Sprintf("./%v/%v", "logs", logName)) // just pass the file name
	if err == nil {
		response.Log = string(logData)
	}
	if logFile.HasNote {
		note, err := ioutil.ReadFile(fmt.Sprintf("./%v/%v", "notes", logName)) // just pass the file name
		if err == nil {
			response.Note = string(note)
		}
	}

	return response
}
