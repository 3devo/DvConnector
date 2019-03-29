package routing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/3devo/dvconnector/routing/responses"
	"github.com/tidwall/gjson"

	"github.com/3devo/dvconnector/models"
	"github.com/3devo/dvconnector/utils"
	"github.com/julienschmidt/httprouter"
)

// swagger:route GET /charts/ Charts GetAllCharts
//
// Handler to retrieve all available charts
//
// Returns all charts
//
// Produces:
// 	application/json
// Responses:
//	200: body:[]Chart
func GetAllCharts(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		charts := make([]models.Chart, 0)
		query, _ := utils.QueryBuilder(env, r)

		w.WriteHeader(http.StatusOK)
		query.Find(&charts)
		json.NewEncoder(w).Encode(charts)
	}
}

// swagger:route GET /charts/{uuid} Charts GetChart
//
// Handler to retrieve a single chart
//
// Returns a single chart
//
// Produces:
// 	application/json
// Responses:
//	200: body:Chart
func GetChart(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		Chart := models.Chart{}
		uuid := ps.ByName("uuid")

		if err := env.Db.One("UUID", uuid, &Chart); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Charts",
				"GET",
				err.Error(),
				w)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Chart)
	}
}

// swagger:route POST /charts Charts CreateChart
//
// Handler to create a chart
//
// Creates a new chart
// Produces:
// 	application/json
// Responses:
//	200: ResourceStatusResponse
func CreateChart(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		validation := responses.ChartCreationBody{}
		body, _ := ioutil.ReadAll(r.Body)
		data := gjson.Parse(string(body))

		json.Unmarshal(body, &validation.Data)

		if err := env.Validator.Struct(validation); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Charts",
				"CREATE",
				err.Error(),
				w)
			return
		}
		if env.Db.One("UUID", data.Get("uuid").String(), &models.Chart{}) == nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Charts",
				"CREATE",
				fmt.Sprintf("Chart with %v already exists", data.Get("uuid").String()),
				w)
			return
		}

		if err := env.Db.Save(&validation.Data); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Charts",
				"CREATE",
				err.Error(),
				w)
			return
		} else {
			responses.WriteResourceStatusResponse(
				http.StatusOK,
				"Charts",
				"CREATE",
				"",
				w)
		}
	}
}

// swagger:route PUT /charts/{uuid} Charts UpdateChart
//
// Handler to update a chart
//
// Replaces an existing chart with new values
// Produces:
// 	application/json
// Responses:
//	200: ResourceStatusResponse
func UpdateChart(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		validation := responses.ChartCreationBody{}
		body, _ := ioutil.ReadAll(r.Body)
		data := gjson.Parse(string(body))

		json.Unmarshal(body, &validation.Data)

		if err := env.Validator.Struct(validation); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Charts",
				"UPDATE",
				err.Error(),
				w)
			return
		}

		if err := env.Db.One("UUID", data.Get("uuid").String(), &models.Chart{}); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Charts",
				"UPDATE",
				err.Error(),
				w)
			return
		}

		if err := env.Db.Update(&validation.Data); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusConflict,
				"Charts",
				"UPDATE",
				err.Error(),
				w)
		} else {
			responses.WriteResourceStatusResponse(
				http.StatusOK,
				"Charts",
				"UPDATE",
				"",
				w)
		}
	}
}

// swagger:route DELETE /charts/{uuid} Charts DeleteChart
//
// Handler to delete a chart
//
// Deletes a existing chart
// Produces:
// 	application/json
// Responses:
//	200: ResourceStatusResponse
func DeleteChart(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		Chart := models.Chart{}
		uuid := ps.ByName("uuid")

		if err := env.Db.One("UUID", uuid, &Chart); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Charts",
				"DELETE",
				err.Error(),
				w)
			return
		}

		if err := env.Db.DeleteStruct(&Chart); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Charts",
				"DELETE",
				err.Error(),
				w)
			return
		}
		responses.WriteResourceStatusResponse(
			http.StatusOK,
			"Charts",
			"DELETE",
			"",
			w)

	}
}
