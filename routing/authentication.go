package routing

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/tidwall/gjson"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/routing/responses"
	"github.com/3devo/feconnector/utils"
	"github.com/julienschmidt/httprouter"
)

const (
	ErrorInvalidPassword = "Invalid password"
)

// swagger:route POST /login Authentication Login
//
// Handler to login
//
// Returns a JWT on success
//
// Produces:
// 	application/json
// Responses:
//	200: body:LoginSuccessResponse
func Login(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		user := models.User{}
		parameters := responses.LoginParameters{}
		body, _ := ioutil.ReadAll(r.Body)
		data := gjson.ParseBytes(body)
		json.Unmarshal(body, &parameters)
		if err := env.Validator.Struct(parameters); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Authentication",
				"LOGIN",
				err.Error(),
				w)
			return
		}
		if err := env.Db.One("Username", parameters.Username, &user); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusUnauthorized,
				"Authentication",
				"LOGIN",
				errors.New(ErrorInvalidPassword).Error(),
				w)
			return
		}
		if !utils.CheckPasswordHash(parameters.Password, user.Password) {
			responses.WriteResourceStatusResponse(
				http.StatusUnauthorized,
				"Authentication",
				"LOGIN",
				errors.New(ErrorInvalidPassword).Error(),
				w)
			return
		}
		response := responses.LoginSuccess{}
		expiration := time.Now().Add(time.Hour * time.Duration(1)).Unix()
		if data.Get("rememberMe").Bool() {
			expiration = time.Now().Add(time.Hour * time.Duration(24*30)).Unix()
		}
		token, _ := utils.GenerateJWTToken(user.UUID, expiration)
		response.Data.Token = token
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response.Data)
	}
}

// Logout adds the current token to the blacklist so it can't be used again
// swagger:route POST /logout Authentication logout
//
// Handler to logout
//
// Returns OK
//
// Produces:
// 	application/text
// Responses:
//	200:
func Logout(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		regex := regexp.MustCompile("bearer (.*)")
		token := regex.FindStringSubmatch(r.Header.Get("Authorization"))
		if len(token) > 0 {

			env.Db.Save(&models.BlackListedToken{
				Token:      token[1],
				Expiration: r.Context().Value("expiration").(int64)})
		}
		responses.WriteResourceStatusResponse(
			http.StatusOK,
			"Authentication",
			"LOGOUT",
			"",
			w)
	}
}
