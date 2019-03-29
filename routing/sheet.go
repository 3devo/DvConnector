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

// swagger:route GET /sheets Sheets GetAllSheets
//
// Handler to retrieve all sheets
//
// This will return all available sheets
//
// Produces:
//	application/json
//
// Responses:
//        200: body:[]SheetResponse
func GetAllSheets(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		sheets := make([]models.Sheet, 0)
		query, _ := utils.QueryBuilder(env, r)

		query.Find(&sheets)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(sheets)
	}
}

// swagger:route GET /sheets/{uuid} Sheets GetSheet
//
// Handler to retrieve a single sheets
//
// This will return a single sheet
//
// Produces:
//	application/json
//
// Responses:
//        200: body:SheetResponse
func GetSheet(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		sheet := models.Sheet{}
		uuid := ps.ByName("uuid")

		if err := env.Db.One("UUID", uuid, &sheet); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Sheets",
				"GET",
				err.Error(),
				w)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responses.GenerateSheetResponseObject(&sheet, env))
	}
}

// swagger:route POST /sheets/ Sheets CreateSheet
//
// Handler to create a new sheet object
//
// This will add a new sheet to the database
// Only charts that exist can be added
//
// Produces:
//	application/json
//
// Responses:
//        200: ResourceStatusResponse
func CreateSheet(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		validation := responses.SheetCreationBody{}
		body, _ := ioutil.ReadAll(r.Body)
		data := gjson.Parse(string(body))
		json.Unmarshal(body, &validation.Data)

		if err := env.Validator.Struct(validation); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Sheets",
				"CREATE",
				err.Error(),
				w)
			return
		}
		if env.Db.One("UUID", data.Get("uuid").String(), &models.Sheet{}) == nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Sheets",
				"CREATE",
				fmt.Sprintf("Sheet with %v already exists", data.Get("uuid").String()),
				w)
			return
		}

		if err := env.Db.Save(&validation.Data); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusConflict,
				"Sheets",
				"CREATE",
				err.Error(),
				w)
		} else {
			responses.WriteResourceStatusResponse(
				http.StatusOK,
				"Sheets",
				"CREATE",
				"",
				w)
		}
	}
}

// swagger:route PUT /sheets/{uuid} Sheets UpdateSheet
//
// Handler to update an existing sheet object
//
// This will update an existing sheet object
// Only charts that exist can be added
//
// Produces:
//	application/json
//
// Responses:
//        200: ResourceStatusResponse
func UpdateSheet(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		validation := responses.SheetCreationBody{}
		body, _ := ioutil.ReadAll(r.Body)
		uuid := ps.ByName("uuid")

		json.Unmarshal(body, &validation.Data)

		if err := env.Validator.Struct(validation); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Sheets",
				"UPDATE",
				err.Error(),
				w)
			return
		}
		if err := env.Db.One("UUID", uuid, &models.Sheet{}); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Sheets",
				"UPDATE",
				err.Error(),
				w)
			return
		}

		if err := env.Db.Update(&validation.Data); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusConflict,
				"Sheets",
				"UPDATE",
				err.Error(),
				w)
		} else {
			responses.WriteResourceStatusResponse(
				http.StatusOK,
				"Sheets",
				"UPDATE",
				"",
				w)
			return
		}
	}
}

// swagger:route DELETE /sheets/{uuid} Sheets DeleteSheet
//
// Handler to delete a existing sheet object
//
// This will add delete a sheet from the database
//
// Produces:
//	application/json
//
// Responses:
//        200: ResourceStatusResponse
func DeleteSheet(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		sheet := models.Sheet{}
		uuid := ps.ByName("uuid")

		if err := env.Db.One("UUID", uuid, &sheet); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Sheets",
				"DELETE",
				err.Error(),
				w)
			return
		}
		if err := env.Db.DeleteStruct(&sheet); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Sheets",
				"DELETE",
				err.Error(),
				w)
			return
		}
		responses.WriteResourceStatusResponse(
			http.StatusOK,
			"Sheets",
			"DELETE",
			"",
			w)
	}
}
