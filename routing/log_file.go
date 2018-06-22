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
func GetLogFile(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var logFile models.LogFile
		uuid := ps.ByName("uuid")
		err := env.Db.One("ID", uuid, &logFile)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(logFile)
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
		var logFile models.LogFile
		body, _ := ioutil.ReadAll(r.Body)
		bodyString := string(body)
		uuid := ps.ByName("uuid")
		err := env.Db.One("UUID", uuid, &logFile)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		err = env.Db.Update(&models.LogFile{
			UUID:      uuid,
			Name:      gjson.Get(bodyString, "name").String(),
			Timestamp: gjson.Get(bodyString, "timestamp").Int(),
			HasNote:   gjson.Get(bodyString, "hasNote").Bool(),
		})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Updated LogFile with ID %v without error", uuid)
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
		err = env.Db.DeleteStruct(&logFile)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Removed LogFile with %v successfully", uuid)
	}
}
