package routing

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/3devo/feconnector/routing/responses"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/utils"
	"github.com/julienschmidt/httprouter"
)

// swagger:route POST /users Users CreateUser
//
// Handler to create a user
//
// Creates a new user
// Produces:
// 	application/json
// Responses:
//	200: ResourceStatusResponse
func CreateUser(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		validation := responses.UserCreationBody{}
		body, _ := ioutil.ReadAll(r.Body)

		json.Unmarshal(body, &validation.Data)
		validation.Data.UUID = uuid.New().String()
		validation.Data.Password = utils.HashPassword(validation.Data.Password)
		if err := env.Validator.Struct(validation.Data); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Users",
				"CREATE",
				err.Error(),
				w)
			return
		}

		var users []models.User
		env.Db.All(&users)
		if len(users) > 0 {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Users",
				"CREATE",
				"Only one user allowed",
				w)
			return
		}

		if err := env.Db.Save(&validation.Data); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Users",
				"CREATE",
				err.Error(),
				w)
			return
		} else {
			responses.WriteResourceStatusResponse(
				http.StatusOK,
				"Users",
				"CREATE",
				"",
				w)
			env.HasAuth = true
		}
	}
}

// swagger:route PUT /users/{uuid} Users UpdateUser
//
// Handler to update a user
//
// Replaces an existing user with new values
// Produces:
// 	application/json
// Responses:
//	200: ResourceStatusResponse
func UpdateUser(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		validation := responses.UserCreationBody{}
		body, _ := ioutil.ReadAll(r.Body)
		data := gjson.Parse(string(body))

		json.Unmarshal(body, &validation.Data)

		if err := env.Validator.Struct(validation.Data); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Users",
				"UPDATE",
				err.Error(),
				w)
			return
		}

		if err := env.Db.One("UUID", data.Get("uuid").String(), &models.User{}); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Users",
				"UPDATE",
				err.Error(),
				w)
			return
		}

		if err := env.Db.Update(&validation.Data); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusConflict,
				"Users",
				"UPDATE",
				err.Error(),
				w)
		} else {
			responses.WriteResourceStatusResponse(
				http.StatusOK,
				"Users",
				"UPDATE",
				"",
				w)
		}
	}
}

// swagger:route DELETE /users/{uuid} Users DeleteUser
//
// Handler to delete a user
//
// Deletes a existing user
// Produces:
// 	application/json
// Responses:
//	200: ResourceStatusResponse
func DeleteUser(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		User := models.User{}
		uuid := ps.ByName("uuid")

		if err := env.Db.One("UUID", uuid, &User); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusNotFound,
				"Users",
				"DELETE",
				err.Error(),
				w)
			return
		}

		if err := env.Db.DeleteStruct(&User); err != nil {
			responses.WriteResourceStatusResponse(
				http.StatusInternalServerError,
				"Users",
				"DELETE",
				err.Error(),
				w)
			return
		}
		responses.WriteResourceStatusResponse(
			http.StatusOK,
			"Users",
			"DELETE",
			"",
			w)

	}
}

//   "jti": "f607791d-2d50-4e50-a164-42a799c9c0d0"
