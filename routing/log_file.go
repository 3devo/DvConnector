package routing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/routing/responses"
	"github.com/3devo/feconnector/utils"
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
		w.WriteHeader(http.StatusOK)
		query.Find(&logFiles)
		for _, file := range logFiles {
			body = append(body, responses.GenerateLogResponse(&file, env))
		}
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
//	404: StatusResponse
func GetLogFile(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var logFile models.LogFile
		uuid := ps.ByName("uuid")
		err := env.Db.One("ID", uuid, &logFile)
		if err != nil {
			responses.WriteStatusResponse(
				http.StatusNotFound,
				fmt.Sprintf("LogFile with uuid:%v not found", uuid),
				w)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(logFile)
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
//	default: StatusResponse
func CreateLogFile(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var validateModel responses.LogFileUpdateBody
		body, _ := ioutil.ReadAll(r.Body)
		data := gjson.Parse(string(body))
		//validation
		json.Unmarshal([]byte(data.String()), &validateModel.Data)
		err := env.Validator.Struct(validateModel)
		if err != nil {
			responses.WriteStatusResponse(
				http.StatusInternalServerError,
				fmt.Sprintf("Log with uuid %v failed to create with error [%v]", data.Get("uuid").String(), err),
				w)
			return
		}

		if env.Db.One("UUID", data.Get("uuid").String(), &models.LogFile{}) == nil {
			responses.WriteStatusResponse(
				http.StatusInternalServerError,
				fmt.Sprintf("Log with uuid: %v failed to create with error [UUID already exists]", data.Get("uuid").String()),
				w)
			return
		}
		logFile, _ := models.CreateLogFile(
			data.Get("uuid").String(),
			data.Get("name").String(),
			data.Get("note").String())
		err = env.Db.Save(logFile)
		if err != nil {
			responses.WriteStatusResponse(
				http.StatusInternalServerError,
				fmt.Sprintf("Log with uuid: %v failed to create with error [%v]", logFile.UUID, err.Error()),
				w)
			return
		}
		responses.WriteStatusResponse(
			http.StatusOK,
			fmt.Sprintf("Log with uuid: %v has been successfully created", logFile.UUID),
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
//	default: StatusResponse
func UpdateLogFile(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var validateModel responses.LogFileUpdateBody
		var logFile models.LogFile
		body, _ := ioutil.ReadAll(r.Body)
		data := gjson.Parse(string(body))
		//validation
		json.Unmarshal(body, &validateModel.Data)
		err := env.Validator.Struct(validateModel)
		if err != nil {
			responses.WriteStatusResponse(
				http.StatusInternalServerError,
				fmt.Sprintf("Log with uuid %v failed to create with error [%v]", data.Get("uuid").String(), err),
				w)
			return
		}
		uuid := ps.ByName("uuid")
		err = env.Db.One("UUID", uuid, &logFile)
		if err != nil {
			responses.WriteStatusResponse(
				http.StatusNotFound,
				fmt.Sprintf("LogFile with uuid:%v not found", data.Get("uuid").String()),
				w)
			return
		}
		logFile.UpdateLogFile(
			data.Get("name").String(),
			data.Get("note").String())
		err = env.Db.Update(&logFile)
		if err != nil {
			responses.WriteStatusResponse(
				http.StatusConflict,
				fmt.Sprintf("Failed to update log with error [%v]", err.Error()),
				w)
		} else {
			responses.WriteStatusResponse(
				http.StatusOK,
				fmt.Sprintf("Updated LogFile with ID %v without error", uuid),
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
//	default: StatusResponse
func DeleteLogFile(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var logFile models.LogFile
		uuid := ps.ByName("uuid")
		err := env.Db.One("UUID", uuid, &logFile)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		err = logFile.DeleteLogFile(env)
		if err != nil {
			responses.WriteStatusResponse(
				http.StatusConflict,
				fmt.Sprintf("Log with uuid: %v failed to create with error [%v]", logFile.UUID, err.Error()),
				w)
			return
		}
		responses.WriteStatusResponse(
			http.StatusOK,
			fmt.Sprintf("Log with uuid: %v has been successfully deleted", logFile.UUID),
			w)
	}
}
