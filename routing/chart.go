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

func GetAllCharts(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		charts := make([]models.Chart, 0)
		query, _ := utils.QueryBuilder(env, r)
		w.WriteHeader(http.StatusOK)
		query.Find(&charts)
		json.NewEncoder(w).Encode(charts)
	}
}

func GetChart(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var Chart models.Chart
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "id should be a number", http.StatusNotAcceptable)
			return
		}
		err = env.Db.One("ID", id, &Chart)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Chart)
	}
}

func CreateChart(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var chart models.Chart
		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &chart)
		err = env.Db.Save(&chart)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Added chart %v without error", chart.ID)
		}
	}
}

func UpdateChart(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var chart models.Chart
		body, _ := ioutil.ReadAll(r.Body)
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "id should be a number", http.StatusNotAcceptable)
			return
		}
		chart.ID = id
		json.Unmarshal(body, &chart)
		err = env.Db.One("ID", id, &models.Chart{})
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		err = env.Db.Update(&chart)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Updated Chart with ID %v without error", id)
		}
	}
}

func DeleteChart(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var Chart models.Chart
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "id should be a number", http.StatusNotAcceptable)
			return
		}
		err = env.Db.One("ID", id, &Chart)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		err = env.Db.DeleteStruct(&Chart)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Deleted Chart with ID %v without error", id)

	}
}
