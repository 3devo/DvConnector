package routing

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/utils"
	"github.com/julienschmidt/httprouter"
	"github.com/tidwall/gjson"
)

func Login(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		session, err := env.SessionStore.Get(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var config models.Config
		err = env.Db.One("ID", 1, &config)
		if err != nil {
			http.Error(w, "config "+err.Error(), http.StatusInternalServerError)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		password := gjson.Get(string(body), "password").String()
		if !config.CheckPasswordhash(password) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		session.Values["logged_in"] = true
		session.Save(r, w)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Login successful")
		return
	}
}
