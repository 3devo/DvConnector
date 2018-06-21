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

func GetAllSheets(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		sheets := make([]models.Sheet, 0)
		responseObject := make([]models.SheetResponse, 0)
		query, _ := utils.QueryBuilder(env, r)
		query.Find(&sheets)
		for _, sheet := range sheets {
			responseObject = append(responseObject, sheet.GetResponseObject(env))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responseObject)
	}
}

func GetSheet(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var sheet models.Sheet
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "id should be a number", http.StatusNotAcceptable)
			return
		}
		err = env.Db.One("ID", id, &sheet)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(sheet.GetResponseObject(env))
	}
}

func CreateSheet(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var sheet models.Sheet
		body, _ := ioutil.ReadAll(r.Body)

		json.Unmarshal(body, &sheet)
		for _, chartId := range sheet.Charts {
			log.Println(chartId)
			if env.Db.One("ID", chartId, &models.Chart{}) != nil {
				http.Error(w, fmt.Sprintf("Chart with id %v doesn't exist", chartId), http.StatusConflict)
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

func UpdateSheet(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var sheet models.Sheet
		body, _ := ioutil.ReadAll(r.Body)
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "id should be a number", http.StatusNotAcceptable)
			return
		}
		sheet.ID = id
		json.Unmarshal(body, &sheet)
		err = env.Db.One("ID", id, &models.Sheet{})
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		for _, chartId := range sheet.Charts {
			if env.Db.One("ID", chartId, &models.Chart{}) != nil {
				http.Error(w, fmt.Sprintf("Chart with id %v doesn't exist", chartId), http.StatusConflict)
				return
			}
		}
		err = env.Db.Update(&sheet)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Updated Chart with ID %v without error", id)
		}
	}
}

func DeleteSheet(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var Sheet models.Sheet
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "id should be a number", http.StatusNotAcceptable)
			return
		}
		err = env.Db.One("ID", id, &Sheet)
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
		fmt.Fprintf(w, "Removed Sheet with %v successfully", id)
	}
}
