package routing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/utils"
	"github.com/julienschmidt/httprouter"
	"github.com/tidwall/gjson"
)

func GetAllLogFiles(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		logFiles := make([]models.LogFile, 0)
		query, _ := utils.QueryBuilder(env, r)
		w.WriteHeader(http.StatusOK)
		query.Find(&logFiles)
		json.NewEncoder(w).Encode(logFiles)
	}
}

func GetLogFile(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var logFile models.LogFile
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "id should be a number", http.StatusNotAcceptable)
			return
		}
		err = env.Db.One("ID", id, &logFile)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(logFile)
	}
}

func CreateLogFile(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		body, _ := ioutil.ReadAll(r.Body)
		has_note := false
		if gjson.Get(string(body), "note").Exists() {
			has_note = true
			//create note
		}
		logFile := models.NewLogFile(
			gjson.Get(string(body), "name").String(),
			time.Now().Unix(),
			has_note)
		log.Print(logFile)
		err := env.Db.Save(logFile)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Added logFile without error")
		}
	}
}

func UpdateLogFile(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var logFile models.LogFile
		body, _ := ioutil.ReadAll(r.Body)
		bodyString := string(body)
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "id should be a number", http.StatusNotAcceptable)
			return
		}
		err = env.Db.One("ID", id, &logFile)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		err = env.Db.Update(&models.LogFile{
			ID:        id,
			Name:      gjson.Get(bodyString, "name").String(),
			Timestamp: gjson.Get(bodyString, "timestamp").Int(),
			HasNote:   gjson.Get(bodyString, "hasNote").Bool(),
		})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Updated LogFile with ID %v without error", id)
		}
	}
}

func DeleteLogFile(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var logFile models.LogFile
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "id should be a number", http.StatusNotAcceptable)
			return
		}
		err = env.Db.One("ID", id, &logFile)
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
		fmt.Fprintf(w, "Removed LogFile with %v successfully", id)
	}
}
