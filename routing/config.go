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
	"github.com/tidwall/gjson"
)

func GetConfig(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var config models.Config
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "id should be a number", http.StatusNotAcceptable)
			return
		}
		err = env.Db.One("ID", id, &config)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		config.Password = nil
		json.NewEncoder(w).Encode(config)
	}
}

func CreateConfig(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		body, _ := ioutil.ReadAll(r.Body)
		config := models.NewConfig(
			gjson.Get(string(body), "authRequired").Bool(),
			gjson.Get(string(body), "spectatorsAllowed").Bool(),
			gjson.Get(string(body), "password").String())
		log.Print(config)
		err := env.Db.Save(config)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Added config without error")
		}
	}
}
