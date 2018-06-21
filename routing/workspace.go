package routing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/utils"
	"github.com/julienschmidt/httprouter"
)

func GetAllWorkspaces(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		workspaces := make([]models.Workspace, 0)
		responseObject := make([]models.WorkspaceResponse, 0)
		query, _ := utils.QueryBuilder(env, r)
		query.Find(&workspaces)
		for _, Workspace := range workspaces {
			responseObject = append(responseObject, Workspace.GetResponseObject(env))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responseObject)
	}
}

func GetWorkspace(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var Workspace models.Workspace
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "id should be a number", http.StatusNotAcceptable)
			return
		}
		err = env.Db.One("ID", id, &Workspace)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Workspace.GetResponseObject(env))
	}
}

func CreateWorkspace(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var Workspace models.Workspace
		body, _ := ioutil.ReadAll(r.Body)

		json.Unmarshal(body, &Workspace)
		for _, sheetId := range Workspace.Sheets {
			if env.Db.One("ID", sheetId, &models.Sheet{}) != nil {
				http.Error(w, fmt.Sprintf("Sheet with id %v doesn't exist", sheetId), http.StatusConflict)
				return
			}
		}
		err := env.Db.Save(&Workspace)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Added Workspace without error")
		}
	}
}

func UpdateWorkspace(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var Workspace models.Workspace
		body, _ := ioutil.ReadAll(r.Body)
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "id should be a number", http.StatusNotAcceptable)
			return
		}
		Workspace.ID = id
		json.Unmarshal(body, &Workspace)
		err = env.Db.One("ID", id, &models.Workspace{})
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		for _, sheetId := range Workspace.Sheets {
			if env.Db.One("ID", sheetId, &models.Sheet{}) != nil {
				http.Error(w, fmt.Sprintf("Sheet with id %v doesn't exist", sheetId), http.StatusConflict)
				return
			}
		}
		err = env.Db.Update(&Workspace)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Updated Workspace with ID %v without error", id)
		}
	}
}

func DeleteWorkspace(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var Workspace models.Workspace
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "id should be a number", http.StatusNotAcceptable)
			return
		}
		err = env.Db.One("ID", id, &Workspace)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		err = env.Db.DeleteStruct(&Workspace)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Removed Workspace with %v successfully", id)
	}
}
