package routing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/3devo/dvconnector/models"
	"github.com/3devo/dvconnector/routing/responses"
	"github.com/3devo/dvconnector/utils"
	"github.com/julienschmidt/httprouter"
	"github.com/tidwall/gjson"
)

// swagger:route GET /logFiles logFiles GetAllLogFiles
//
// Handler to retrieve all logFiles
//
// This will return all available logs
//
// Produces:
//	application/json
//
// Responses:
//        200: []LogFileResponse
func GetAllLogFiles(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		logFiles := make([]models.LogFile, 0)
		body := make([]*responses.LogFileResponse, 0)
		query, _ := utils.QueryBuilder(env, r)

		query.Find(&logFiles)
		for _, file := range logFiles {
			body = append(body, responses.GenerateLogResponse(&file, env))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(body)
	}
}

// swagger:route GET /logFiles/{uuid} logFiles GetLogFile
//
// Handler to retrieve a single logFile
//
// This will return a single log
//
// Produces:
//	application/json
//
// Responses:
// 	200: LogFileResponse
//	404: ResourceStatusResponse
func GetLogFile(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		logFile := models.LogFile{}
		uuid := ps.ByName("uuid")

		if err := env.Db.One("UUID", uuid, &logFile); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Logfiles",
				"GET",
				err.Error(),
				w)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responses.GenerateLogResponse(&logFile, env))
	}
}

//swagger:route POST /logFiles logFiles CreateLogFile
//
// Handler to create a new log file
//
// This method will create a new log file in the database
// and create new physical files in the logs, notes directory.
//
// Produces:
// 	application/json
//
// Consumes:
// 	application/json
//
// Responses:
//	default: ResourceStatusResponse
func CreateLogFile(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		validateModel := responses.LogFileCreationBody{}
		body, _ := ioutil.ReadAll(r.Body)
		data := gjson.Parse(string(body))

		json.Unmarshal([]byte(data.String()), &validateModel.Data)
		if err := env.Validator.Struct(validateModel); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Logfiles",
				"CREATE",
				err.Error(),
				w)
			return
		}
		if env.Db.One("UUID", data.Get("uuid").String(), &models.LogFile{}) == nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Logfiles",
				"CREATE",
				fmt.Sprintf("Logfile with %v already exists", data.Get("uuid").String()),
				w)
			return
		}
		err := models.CreateLogFile(
			data.Get("uuid").String(),
			data.Get("name").String(),
			data.Get("note").String(),
			env)
		if err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Logfiles",
				"CREATE",
				err.Error(),
				w)
			return
		}
		responses.WriteResourceStatusResponse(
			http.StatusOK,
			"Logfiles",
			"CREATE",
			"",
			w)
	}
}

// swagger:route PUT /logFiles/{uuid} logFiles UpdateLogFile
//
// Handler to update the logFile name
//
// This will allow updating of the log name
//
// Consumes:
//	application/json
// Responses:
//	default: ResourceStatusResponse
func UpdateLogFile(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		validateModel := responses.LogFileCreationBody{}
		logFile := models.LogFile{}
		body, _ := ioutil.ReadAll(r.Body)
		data := gjson.Parse(string(body))
		uuid := ps.ByName("uuid")

		json.Unmarshal(body, &validateModel.Data)

		if err := env.Validator.Struct(validateModel); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Logfiles",
				"UPDATE",
				err.Error(),
				w)
			return
		}

		if err := env.Db.One("UUID", uuid, &logFile); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Logfiles",
				"UPDATE",
				err.Error(),
				w)
			return
		}
		err := logFile.UpdateLogFile(
			data.Get("name").String(),
			data.Get("note").String(),
			env)
		if err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusConflict,
				"Logfiles",
				"UPDATE",
				err.Error(),
				w)
		} else {
			responses.WriteResourceStatusResponse(
				http.StatusOK,
				"Logfiles",
				"UPDATE",
				"",
				w)
		}
	}
}

// swagger:route DELETE /logFiles/{uuid} logFiles DeleteLogFile
//
// Handler to delete the logFile name
//
// This will delete a log file
//
// Consumes:
//	application/json
// Responses:
//	default: ResourceStatusResponse
func DeleteLogFile(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		logFile := models.LogFile{}
		uuid := ps.ByName("uuid")

		if err := env.Db.One("UUID", uuid, &logFile); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Logfiles",
				"DELETE",
				err.Error(),
				w)
			return
		}
		if err := logFile.DeleteLogFile(env); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusConflict,
				"Logfiles",
				"DELETE",
				err.Error(),
				w)
			return
		}
		responses.WriteResourceStatusResponse(
			http.StatusOK,
			"Logfiles",
			"DELETE",
			"",
			w)
	}
}
