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
)

// swagger:route GET /workspaces Workspaces GetAllWorkspaces
//
// Handler to retrieve all workspaces
//
// This will return all available workspaces
//
// Produces:
//	application/json
//
// Responses:
//        200: body:[]WorkspaceResponse
func GetAllWorkspaces(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		workspaces := make([]models.Workspace, 0)
		responseObject := make([]*responses.WorkspaceResponse, 0)
		query, _ := utils.QueryBuilder(env, r)
		query.Find(&workspaces)
		for _, workspace := range workspaces {
			responseObject = append(responseObject, responses.GenerateWorkspaceResponseObject(&workspace, env))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responseObject)
	}
}

// swagger:route GET /workspaces/{uuid} Workspaces GetWorkspace
//
// Handler to retrieve a single workspaces
//
// This will return a single workspace
//
// Produces:
//	application/json
//
// Responses:
//        200: body:WorkspaceResponse
func GetWorkspace(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var workspace models.Workspace
		uuid := ps.ByName("uuid")
		err := env.Db.One("UUID", uuid, &workspace)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responses.GenerateWorkspaceResponseObject(&workspace, env))
	}
}

// swagger:route POST /workspaces Workspaces CreateWorkspace
//
// Handler to create a new workspace object
//
// This will add a new workspace to the database
// Only sheets that exist can be added
//
// Produces:
//	application/json
//
// Responses:
//        200: StatusResponse
func CreateWorkspace(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var Workspace models.Workspace
		body, _ := ioutil.ReadAll(r.Body)

		json.Unmarshal(body, &Workspace)
		for _, sheetId := range Workspace.Sheets {
			if env.Db.One("UUID", sheetId, &models.Sheet{}) != nil {
				http.Error(w, fmt.Sprintf("Sheet with uuid %v doesn't exist", sheetId), http.StatusConflict)
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

// swagger:route PUT /workspaces/{uuid} Workspaces UpdateWorkspace
//
// Handler to update an existing workspace object
//
// This will update an existing workspace object
// Only sheets that exist can be added
//
// Produces:
//	application/json
//
// Responses:
//        200: StatusResponse
func UpdateWorkspace(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var Workspace models.Workspace
		body, _ := ioutil.ReadAll(r.Body)
		uuid := ps.ByName("uuid")
		json.Unmarshal(body, &Workspace)
		err := env.Db.One("UUID", uuid, &models.Workspace{})
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		for _, sheetId := range Workspace.Sheets {
			if env.Db.One("UUID", sheetId, &models.Sheet{}) != nil {
				http.Error(w, fmt.Sprintf("Sheet with uuid %v doesn't exist", sheetId), http.StatusConflict)
				return
			}
		}
		err = env.Db.Update(&Workspace)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Updated Workspace with ID %v without error", uuid)
		}
	}
}

// swagger:route DELETE /workspaces/{uuid} Workspaces DeleteWorkspace
//
// Handler to delete a existing workspace object
//
// This will add delete a workspace from the database
//
// Produces:
//	application/json
//
// Responses:
//        200: StatusResponse
func DeleteWorkspace(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var Workspace models.Workspace
		uuid := ps.ByName("uuid")
		err := env.Db.One("UUID", uuid, &Workspace)
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
		fmt.Fprintf(w, "Removed Workspace with %v successfully", uuid)
	}
}
