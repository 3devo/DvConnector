package routing

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/tidwall/gjson"

	"github.com/3devo/dvconnector/models"
	"github.com/3devo/dvconnector/routing/responses"
	"github.com/3devo/dvconnector/utils"
	"github.com/julienschmidt/httprouter"
)

const (
	ErrorInvalidPassword = "Invalid password" // Invalid password error
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
		expiration := time.Now().Add(time.Minute * time.Duration(utils.StandardTokenExpiration)).Unix()
		if data.Get("rememberMe").Bool() {
			expiration = time.Now().Add(time.Hour * time.Duration(utils.ExtendedTokenExpiration)).Unix()
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
// 	application/json
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

// RefreshToken returns a new token
// swagger:route POST /refreshToken Authentication refresh
//
// Handler that returns a new token if not expired
//
// Produces
//	application/json
// Responses:
//	200: body:LoginSuccessResponse
func RefreshToken(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		response := responses.LoginSuccess{}
		expiration := time.Unix(r.Context().Value("expiration").(int64), 0).Add(time.Minute * time.Duration(utils.StandardTokenExpiration)).Unix()
		token, _ := utils.GenerateJWTToken(r.Context().Value("userId").(string), expiration)
		response.Data.Token = token
		env.Db.Save(&models.BlackListedToken{
			Token:      r.Context().Value("token").(string),
			Expiration: r.Context().Value("expiration").(int64)})
		json.NewEncoder(w).Encode(response.Data)
		w.WriteHeader(http.StatusOK)
	}
}
