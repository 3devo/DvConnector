package routing_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/routing"
	"github.com/3devo/feconnector/routing/responses"
	"github.com/google/uuid"

	"github.com/3devo/feconnector/utils"
	"github.com/julienschmidt/httprouter"
	. "github.com/smartystreets/goconvey/convey"
	validator "gopkg.in/go-playground/validator.v9"
)

func TestCreateUser(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), FileDir: dir}

		updateBody := &responses.UserCreationBody{}
		updateBody.Data.UUID = uuid.New().String()
		updateBody.Data.Username = "test@test.nl"
		updateBody.Data.Password = "test note"

		env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
			return utils.IsValidUUID(fl.Field().String())
		})

		router := httprouter.New()
		Convey("Given a HTTP POST request for api/x/users with a valid body", func() {
			router.POST("/api/x/users", routing.CreateUser(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/users", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a success status and a new user should have been created", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Users"
				response.Body.Action = "CREATE"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/users when a user exists", func() {
			db.Save(&models.User{
				UUID:     uuid.New().String(),
				Username: "bob",
				Password: "password"})
			router.POST("/api/x/users", routing.CreateUser(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/users", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail because there is already a user active", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Users"
				response.Body.Action = "CREATE"
				response.Body.Error = "Only one user allowed"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/users with a missing username in body", func() {
			router.POST("/api/x/users", routing.CreateUser(env))
			updateBody.Data.Username = ""
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/users", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with required username message", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Users"
				response.Body.Action = "CREATE"
				response.Body.Error = "Key: 'User.Username' Error:Field validation for 'Username' failed on the 'required' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestUpdateUser(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		user := models.User{
			UUID:     uuid.New().String(),
			Username: "test@test.nl",
			Password: "password"}

		env := &utils.Env{Db: db, Validator: validator.New(), FileDir: path.Dir(dir)}
		updateBody := &responses.UserCreationBody{}
		updateBody.Data.UUID = user.UUID
		updateBody.Data.Username = user.Username
		updateBody.Data.Password = "password"
		router := httprouter.New()

		env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
			return utils.IsValidUUID(fl.Field().String())
		})
		db.Save(&user)
		Convey("Given a HTTP PUT request for api/x/users/uuid with a valid body", func() {
			router.PUT("/api/x/users/:uuid", routing.UpdateUser(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/users/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a successful", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Users"
				response.Body.Action = "UPDATE"
				response.Body.Error = ""
				expected, _ := json.Marshal(response.Body)

				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/users/uuid with a unknown uid", func() {
			router.PUT("/api/x/users/:uuid", routing.UpdateUser(env))
			updateBody.Data.UUID = "550e8400-e29b-41d4-a716-446655440004"
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/users/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should return a not found error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusNotFound
				response.Body.Resource = "Users"
				response.Body.Action = "UPDATE"
				response.Body.Error = "not found"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusNotFound)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/users/uuid with a invalid uuid in the body", func() {
			router.PUT("/api/x/users/:uuid", routing.UpdateUser(env))
			updateBody.Data.UUID = ""
			requestBody, _ := json.Marshal(updateBody.Data)
			updateBody.Data.UUID = user.UUID
			req := httptest.NewRequest("PUT", "/api/x/users/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should return a internal server error with a uuid validation error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Users"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'User.UUID' Error:Field validation for 'UUID' failed on the 'uuid' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/users/uuid with a missing name", func() {
			router.PUT("/api/x/users/:uuid", routing.UpdateUser(env))
			updateBody.Data.Username = ""

			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/users/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with name validation error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Users"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'User.Username' Error:Field validation for 'Username' failed on the 'required' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestDeleteUser(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		user := models.User{
			UUID:     uuid.New().String(),
			Username: "bob",
			Password: "password"}
		db.Save(&user)

		env := &utils.Env{Db: db, Validator: validator.New(), FileDir: dir}
		router := httprouter.New()
		Convey("Given a HTTP DELETE request for api/x/users/uuid with a known uuid", func() {
			router.DELETE("/api/x/users/:uuid", routing.DeleteUser(env))

			req := httptest.NewRequest("DELETE", "/api/x/users/"+user.UUID, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be successful", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Users"
				response.Body.Action = "DELETE"
				response.Body.Error = ""
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(string(body), ShouldResemble, string(append(expected, 10)))
				So(db.One("Username", "bob", &models.User{}), ShouldNotBeNil)
			})
		})

		Convey("Given a HTTP DELETE request for api/x/users/uuid with a unknown uuid", func() {
			router.DELETE("/api/x/users/:uuid", routing.DeleteUser(env))

			req := httptest.NewRequest("DELETE", "/api/x/users/unknown", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail and return not found", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusNotFound
				response.Body.Resource = "Users"
				response.Body.Action = "DELETE"
				response.Body.Error = "not found"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusNotFound)
				So(string(body), ShouldResemble, string(append(expected, 10)))
				So(db.One("Username", "bob", &models.User{}), ShouldBeNil)
			})
		})
	})
}
