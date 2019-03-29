package routing_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/3devo/dvconnector/models"
	"github.com/3devo/dvconnector/routing"
	"github.com/3devo/dvconnector/routing/responses"

	"github.com/3devo/dvconnector/utils"
	"github.com/julienschmidt/httprouter"
	. "github.com/smartystreets/goconvey/convey"
	validator "gopkg.in/go-playground/validator.v9"
)

func TestGetSingleWorkspace(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir)}
		Convey("Given a HTTP request for /api/x/workspaces/550e8400-e29b-41d4-a716-446655440000", func() {

			router := httprouter.New()
			router.GET("/api/x/workspaces/:uuid", routing.GetWorkspace(env))

			req := httptest.NewRequest("GET", "/api/x/workspaces/550e8400-e29b-41d4-a716-446655440000", nil)
			resp := httptest.NewRecorder()

			Convey("When the request is handled by the Router", func() {
				router.ServeHTTP(resp, req)

				Convey("Then the response should be a http.StatusOK because the item exists in the db", func() {
					result := resp.Result()
					body, _ := ioutil.ReadAll(result.Body)
					expected, _ := json.Marshal(responses.GenerateWorkspaceResponseObject(&workspaces[0], env))
					So(result.StatusCode, ShouldEqual, http.StatusOK)
					So(body, ShouldResemble, append(expected, 10))
				})
			})
		})

		Convey("Given a HTTP request for /api/x/workspaces/undefined", func() {
			router := httprouter.New()
			router.GET("/api/x/workspaces/:uuid", routing.GetWorkspace(env))

			req := httptest.NewRequest("GET", "/api/x/workspaces/undefined", nil)
			resp := httptest.NewRecorder()

			Convey("When the request is handled by the Router", func() {
				router.ServeHTTP(resp, req)

				Convey("Then the response should be a http.StatusNotFound with correct response body", func() {
					result := resp.Result()
					body, _ := ioutil.ReadAll(result.Body)
					response := responses.ResourceStatusResponse{}
					response.Body.Code = http.StatusNotFound
					response.Body.Resource = "Workspaces"
					response.Body.Action = "GET"
					response.Body.Error = "not found"
					expected, _ := json.Marshal(response.Body)

					So(result.StatusCode, ShouldEqual, http.StatusNotFound)
					So(string(body), ShouldResemble, string(append(expected, 10)))
				})
			})
		})
	})
}

func TestGetMultipleWorkspaces(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir)}

		Convey("Given a HTTP request for api/x/workspaces", func() {
			router := httprouter.New()
			router.GET("/api/x/workspaces", routing.GetAllWorkspaces(env))

			req := httptest.NewRequest("GET", "/api/x/workspaces", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 3 workspaces", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				expected, _ := json.Marshal(workspaces)
				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/workspaces?filter=[{\"key\":\"title\", \"value\":\"workspace0\"}]", func() {
			router := httprouter.New()
			router.GET("/api/x/workspaces", routing.GetAllWorkspaces(env))

			req := httptest.NewRequest("GET", "/api/x/workspaces", nil)
			q := req.URL.Query()
			q.Add("filter", "[{\"key\":\"title\", \"value\":\"workspace0\"}]")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 1 workspaces", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				expected, _ := json.Marshal([]models.Workspace{workspaces[0]})

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/workspaces?skip=1", func() {
			router := httprouter.New()
			router.GET("/api/x/workspaces", routing.GetAllWorkspaces(env))

			req := httptest.NewRequest("GET", "/api/x/workspaces", nil)
			q := req.URL.Query()
			q.Add("skip", "1")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 2 workspaces", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				expected, _ := json.Marshal(workspaces[1:])

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP request for api/x/workspaces?limit=1", func() {
			router := httprouter.New()
			router.GET("/api/x/workspaces", routing.GetAllWorkspaces(env))

			req := httptest.NewRequest("GET", "/api/x/workspaces", nil)
			q := req.URL.Query()
			q.Add("limit", "1")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 2 workspaces", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				expected, _ := json.Marshal([]models.Workspace{workspaces[0]})

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/workspaces?orderBy=[title]", func() {
			router := httprouter.New()
			router.GET("/api/x/workspaces", routing.GetAllWorkspaces(env))

			req := httptest.NewRequest("GET", "/api/x/workspaces", nil)
			q := req.URL.Query()
			q.Add("orderBy", "title")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 3 workspaces", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				expected, _ := json.Marshal(workspaces)
				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/workspaces?orderBy=[title]&reverse=true", func() {
			router := httprouter.New()
			router.GET("/api/x/workspaces", routing.GetAllWorkspaces(env))

			req := httptest.NewRequest("GET", "/api/x/workspaces", nil)
			q := req.URL.Query()
			q.Add("orderBy", "title")
			q.Add("reverse", "true")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 3 workspaces", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				reversed := [3]models.Workspace{
					workspaces[2],
					workspaces[1],
					workspaces[0],
				}
				expected, _ := json.Marshal(reversed)
				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})
	})
}

func TestCreateWorkspace(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: dir}

		updateBody := &responses.WorkspaceCreationBody{Data: workspaces[0]}
		updateBody.Data.UUID = "550e8400-e29b-41d4-a716-446655440003"
		env.Validator.RegisterValidation("sheet-exists", func(fl validator.FieldLevel) bool {
			return true
		})
		router := httprouter.New()
		Convey("Given a HTTP POST request for api/x/workspaces with a valid body", func() {
			env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
				return true
			})
			router.POST("/api/x/workspaces", routing.CreateWorkspace(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/workspaces", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a success status and a new workspace should have been created", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Workspaces"
				response.Body.Action = "CREATE"
				expected, _ := json.Marshal(response.Body)

				So(string(body), ShouldResemble, string(append(expected, 10)))
				So(result.StatusCode, ShouldEqual, http.StatusOK)
			})
		})

		Convey("Given a HTTP POST request for api/x/workspaces with an existing uuid", func() {
			env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
				return true
			})
			updateBody.Data.UUID = workspaces[0].UUID
			router.POST("/api/x/workspaces", routing.CreateWorkspace(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/workspaces", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail and return already exists error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Workspaces"
				response.Body.Action = "CREATE"
				response.Body.Error = fmt.Sprintf("Workspace with %v already exists", updateBody.Data.UUID)
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/workspaces with a invalid uuid in body", func() {
			env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
				return false
			})
			router.POST("/api/x/workspaces", routing.CreateWorkspace(env))
			updateBody.Data.UUID = "invalid"
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/workspaces", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with uuid validation fail error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Workspaces"
				response.Body.Action = "CREATE"
				response.Body.Error = "Key: 'WorkspaceCreationBody.Data.UUID' Error:Field validation for 'UUID' failed on the 'uuid' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/workspaces with a missing name in body", func() {
			router.POST("/api/x/workspaces", routing.CreateWorkspace(env))
			updateBody.Data.Title = ""
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/workspaces", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with required name message", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Workspaces"
				response.Body.Action = "CREATE"
				response.Body.Error = "Key: 'WorkspaceCreationBody.Data.Title' Error:Field validation for 'Title' failed on the 'required' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/workspaces with non existant sheet", func() {
			env.Validator.RegisterValidation("sheet-exists", func(fl validator.FieldLevel) bool {
				return false
			})
			router.POST("/api/x/workspaces", routing.CreateWorkspace(env))
			updateBody.Data.Sheets = []string{"unknown"}
			requestBody, _ := json.Marshal(updateBody.Data)

			req := httptest.NewRequest("POST", "/api/x/workspaces", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with required name message", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Workspaces"
				response.Body.Action = "CREATE"
				response.Body.Error = "Key: 'WorkspaceCreationBody.Data.Sheets[0]' Error:Field validation for 'Sheets[0]' failed on the 'sheet-exists' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestUpdateWorkspace(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir)}

		env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
			return utils.IsValidUUID(fl.Field().String())
		})
		env.Validator.RegisterValidation("sheet-exists", func(fl validator.FieldLevel) bool {
			return true
		})
		updateBody := responses.WorkspaceCreationBody{Data: workspaces[0]}

		router := httprouter.New()
		Convey("Given a HTTP PUT request for api/x/workspaces/uuid with a valid body", func() {
			router.PUT("/api/x/workspaces/:uuid", routing.UpdateWorkspace(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/workspaces/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a successful", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Workspaces"
				response.Body.Action = "UPDATE"
				response.Body.Error = ""
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/workspaces/uuid with a unknown uid", func() {
			router.PUT("/api/x/workspaces/:uuid", routing.UpdateWorkspace(env))
			updateBody.Data.UUID = "550e8400-e29b-41d4-a716-446655440004"
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/workspaces/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should return a not found error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusNotFound
				response.Body.Resource = "Workspaces"
				response.Body.Action = "UPDATE"
				response.Body.Error = "not found"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusNotFound)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/workspaces/uuid with a invalid uuid in the body", func() {
			router.PUT("/api/x/workspaces/:uuid", routing.UpdateWorkspace(env))
			updateBody.Data.UUID = ""
			requestBody, _ := json.Marshal(updateBody.Data)
			updateBody.Data.UUID = workspaces[0].UUID
			req := httptest.NewRequest("PUT", "/api/x/workspaces/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should return a internal server error with a uuid validation error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Workspaces"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'WorkspaceCreationBody.Data.UUID' Error:Field validation for 'UUID' failed on the 'uuid' tag"
				expected, _ := json.Marshal(response.Body)

				So(string(body), ShouldResemble, string(append(expected, 10)))
				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Given a HTTP PUT request for api/x/workspaces/uuid with a missing name", func() {
			router.PUT("/api/x/workspaces/:uuid", routing.UpdateWorkspace(env))
			updateBody.Data.Title = ""

			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/workspaces/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with name validation error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Workspaces"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'WorkspaceCreationBody.Data.Title' Error:Field validation for 'Title' failed on the 'required' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/workspaces/uuid with a unknown sheet", func() {
			env.Validator.RegisterValidation("sheet-exists", func(fl validator.FieldLevel) bool {
				return false
			})

			router.PUT("/api/x/workspaces/:uuid", routing.UpdateWorkspace(env))

			updateBody.Data.Sheets = []string{"unknown"}
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/workspaces/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with name validation error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Workspaces"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'WorkspaceCreationBody.Data.Sheets[0]' Error:Field validation for 'Sheets[0]' failed on the 'sheet-exists' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestDeleteWorkspace(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: dir}
		router := httprouter.New()
		Convey("Given a HTTP DELETE request for api/x/workspaces/uuid with a known uuid", func() {
			router.DELETE("/api/x/workspaces/:uuid", routing.DeleteWorkspace(env))

			req := httptest.NewRequest("DELETE", "/api/x/workspaces/"+workspaces[0].UUID, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be successful", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Workspaces"
				response.Body.Action = "DELETE"
				response.Body.Error = ""
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP DELETE request for api/x/workspaces/uuid with a unknown uuid", func() {
			router.DELETE("/api/x/workspaces/:uuid", routing.DeleteWorkspace(env))

			req := httptest.NewRequest("DELETE", "/api/x/workspaces/unknown", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail and return not found", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusNotFound
				response.Body.Resource = "Workspaces"
				response.Body.Action = "DELETE"
				response.Body.Error = "not found"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusNotFound)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}
