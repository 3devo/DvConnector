package routing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/3devo/feconnector/routing/responses"
	"github.com/tidwall/gjson"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/utils"
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
		var Chart models.Chart
		uuid := ps.ByName("uuid")
		err := env.Db.One("UUID", uuid, &Chart)
		if err != nil {
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
		var validation responses.ChartCreationParam
		body, _ := ioutil.ReadAll(r.Body)
		data := gjson.Parse(string(body))

		json.Unmarshal(body, &validation.Data)
		err := env.Validator.Struct(validation)
		if err != nil {
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
				err.Error(),
				w)
			return
		}

		err = env.Db.Save(&validation.Data)
		if err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Charts",
				"CREATE",
				err.Error(),
				w)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Added chart %v without error", validation.Data.UUID)
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
		var chart models.Chart
		body, _ := ioutil.ReadAll(r.Body)
		uuid := ps.ByName("uuid")
		chart.UUID = uuid
		json.Unmarshal(body, &chart)
		err := env.Db.One("UUID", uuid, &models.Chart{})
		if err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Charts",
				"UPDATE",
				err.Error(),
				w)
			return
		}

		err = env.Db.Update(&chart)
		if err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusConflict,
				"Charts",
				"UPDATE",
				err.Error(),
				w)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Updated Chart with ID %v without error", uuid)
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
		var Chart models.Chart
		uuid := ps.ByName("uuid")
		err := env.Db.One("UUID", uuid, &Chart)
		if err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Charts",
				"DELETE",
				err.Error(),
				w)
			return
		}
		err = env.Db.DeleteStruct(&Chart)
		if err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Charts",
				"DELETE",
				err.Error(),
				w)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Deleted Chart with ID %v without error", uuid)

	}
}
