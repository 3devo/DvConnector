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
		query, _ := utils.QueryBuilder(env, r)

		query.Find(&workspaces)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(workspaces)
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
		workspace := models.Workspace{}
		uuid := ps.ByName("uuid")
		err := env.Db.One("UUID", uuid, &workspace)

		if err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Workspaces",
				"GET",
				err.Error(),
				w)
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
//        200: ResourceStatusResponse
func CreateWorkspace(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		validation := responses.WorkspaceCreationBody{}
		body, _ := ioutil.ReadAll(r.Body)
		data := gjson.Parse(string(body))

		json.Unmarshal(body, &validation.Data)
		if err := env.Validator.Struct(validation); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Workspaces",
				"CREATE",
				err.Error(),
				w)
			return
		}
		if env.Db.One("UUID", data.Get("uuid").String(), &models.Sheet{}) == nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Workspaces",
				"CREATE",
				fmt.Sprintf("Workspace with %v already exists", data.Get("uuid").String()),
				w)
			return
		}
		if err := env.Db.Save(&validation.Data); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusConflict,
				"Workspaces",
				"CREATE",
				err.Error(),
				w)
		} else {
			responses.WriteResourceStatusResponse(
				http.StatusOK,
				"Workspaces",
				"CREATE",
				"",
				w)
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
//        200: ResourceStatusResponse
func UpdateWorkspace(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		validation := responses.WorkspaceCreationBody{}
		body, _ := ioutil.ReadAll(r.Body)
		uuid := ps.ByName("uuid")

		json.Unmarshal(body, &validation.Data)
		if err := env.Validator.Struct(validation); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Workspaces",
				"UPDATE",
				err.Error(),
				w)
			return
		}
		if err := env.Db.One("UUID", uuid, &models.Workspace{}); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Workspaces",
				"UPDATE",
				err.Error(),
				w)
			return
		}
		if err := env.Db.Update(&validation.Data); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusConflict,
				"Workspaces",
				"UPDATE",
				err.Error(),
				w)
		} else {
			responses.WriteResourceStatusResponse(
				http.StatusOK,
				"Workspaces",
				"UPDATE",
				"",
				w)
			return
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
//        200: ResourceStatusResponse
func DeleteWorkspace(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		workspace := models.Workspace{}
		uuid := ps.ByName("uuid")

		if err := env.Db.One("UUID", uuid, &workspace); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Workspaces",
				"DELETE",
				err.Error(),
				w)
			return
		}

		if err := env.Db.DeleteStruct(&workspace); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Workspaces",
				"DELETE",
				err.Error(),
				w)
			return
		}
		responses.WriteResourceStatusResponse(
			http.StatusOK,
			"Workspaces",
			"DELETE",
			"",
			w)
	}
}
