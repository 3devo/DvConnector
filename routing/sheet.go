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
		responseObject := make([]*responses.SheetResponse, 0)
		query, _ := utils.QueryBuilder(env, r)
		query.Find(&sheets)
		for _, sheet := range sheets {
			responseObject = append(responseObject, responses.GenerateSheetResponseObject(&sheet, env))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responseObject)
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
		var sheet models.Sheet
		uuid := ps.ByName("uuid")
		err := env.Db.One("UUID", uuid, &sheet)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
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
//        200: StatusResponse
func CreateSheet(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var sheet models.Sheet
		body, _ := ioutil.ReadAll(r.Body)

		json.Unmarshal(body, &sheet)
		for _, chartId := range sheet.Charts {
			log.Println(chartId)
			if env.Db.One("UUID", chartId, &models.Chart{}) != nil {
				http.Error(w, fmt.Sprintf("Chart with uuid %v doesn't exist", chartId), http.StatusConflict)
				return
			}
		}
		err := env.Db.Save(&sheet)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Added Sheet without error")
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
//        200: StatusResponse
func UpdateSheet(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var sheet models.Sheet
		body, _ := ioutil.ReadAll(r.Body)
		uuid := ps.ByName("uuid")
		sheet.UUID = uuid
		json.Unmarshal(body, &sheet)
		err := env.Db.One("UUID", uuid, &models.Sheet{})
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		for _, chartId := range sheet.Charts {
			if env.Db.One("ID", chartId, &models.Chart{}) != nil {
				http.Error(w, fmt.Sprintf("Chart with uuid %v doesn't exist", chartId), http.StatusConflict)
				return
			}
		}
		err = env.Db.Update(&sheet)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Updated Chart with ID %v without error", uuid)
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
//        200: StatusResponse
func DeleteSheet(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var Sheet models.Sheet
		uuid := ps.ByName("uuid")
		err := env.Db.One("UUID", uuid, &Sheet)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		err = env.Db.DeleteStruct(&Sheet)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Removed Sheet with %v successfully", uuid)
	}
}
