package routing

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/3devo/dvconnector/models"
	"github.com/3devo/dvconnector/routing/responses"
	"github.com/3devo/dvconnector/utils"
	"github.com/julienschmidt/httprouter"
)

// swagger:route GET /Configs/{uuid} Configs GetConfig
//
// Handler to retrieve the config
//
// This will the config data
//
// Produces:
//	application/json
//
// Responses:
// 	200: ConfigResponse
//	404: ResourceStatusResponse
func GetConfig(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		config := models.Config{}

		if err := env.Db.One("ID", 1, &config); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Configs",
				"GET",
				err.Error(),
				w)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&config)
	}
}

// swagger:route PUT /Config/{uuid} Config UpdateConfig
//
// Handler to update an existing Config object
//
// This will update an existing Config object
//
// Produces:
//	application/json
//
// Responses:
//        200: ResourceStatusResponse
func UpdateConfig(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		validation := responses.ConfigCreationBody{}
		body, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &validation.Data)

		if err := env.Validator.Struct(validation); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Config",
				"UPDATE",
				err.Error(),
				w)
			return
		}
		if err := env.Db.One("ID", 1, &models.Config{}); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Config",
				"UPDATE",
				err.Error(),
				w)
			return
		}

		if err := env.Db.Save(&validation.Data); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusConflict,
				"Config",
				"UPDATE",
				err.Error(),
				w)
		} else {
			responses.WriteResourceStatusResponse(
				http.StatusOK,
				"Config",
				"UPDATE",
				"",
				w)
			return
		}
	}
}
