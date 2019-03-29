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

	"github.com/3devo/dvconnector/routing"
	"github.com/3devo/dvconnector/routing/responses"
	"github.com/google/uuid"

	"github.com/3devo/dvconnector/utils"
	"github.com/julienschmidt/httprouter"
	. "github.com/smartystreets/goconvey/convey"
	validator "gopkg.in/go-playground/validator.v9"
)

func TestGetSingleLogFile(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir)}
		Convey("Given a HTTP request for /api/x/logFiles/550e8400-e29b-41d4-a716-446655440000", func() {

			router := httprouter.New()
			router.GET("/api/x/logFiles/:uuid", routing.GetLogFile(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles/550e8400-e29b-41d4-a716-446655440000", nil)
			resp := httptest.NewRecorder()

			Convey("When the request is handled by the Router", func() {
				router.ServeHTTP(resp, req)

				Convey("Then the response should be a http.StatusOK because the item exists in the db", func() {
					result := resp.Result()
					body, _ := ioutil.ReadAll(result.Body)
					expected, _ := json.Marshal(responses.GenerateLogResponse(&logFiles[0], env))

					So(result.StatusCode, ShouldEqual, http.StatusOK)
					So(body, ShouldResemble, append(expected, 10))
				})
			})
		})

		Convey("Given a HTTP request for /api/x/logFiles/undefined", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles/:uuid", routing.GetLogFile(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles/undefined", nil)
			resp := httptest.NewRecorder()

			Convey("When the request is handled by the Router", func() {
				router.ServeHTTP(resp, req)

				Convey("Then the response should be a http.StatusNotFound with correct response body", func() {
					result := resp.Result()
					body, _ := ioutil.ReadAll(result.Body)
					response := responses.ResourceStatusResponse{}
					response.Body.Code = http.StatusNotFound
					response.Body.Resource = "Logfiles"
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

func TestGetMultipleLogFile(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir)}

		Convey("Given a HTTP request for api/x/logFiles", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles", routing.GetAllLogFiles(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 3 logFiles", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				logResponses := make([]*responses.LogFileResponse, 0)
				for _, model := range logFiles {
					logResponses = append(logResponses, responses.GenerateLogResponse(&model, env))
				}
				expected, _ := json.Marshal(logResponses)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/logFiles?filter=[{\"key\":\"name\", \"value\":\"log1\"}]", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles", routing.GetAllLogFiles(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles", nil)
			q := req.URL.Query()
			q.Add("filter", "[{\"key\":\"name\", \"value\":\"log1\"}]")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 1 logFiles", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				logResponses := make([]*responses.LogFileResponse, 0)
				logResponses = append(logResponses, responses.GenerateLogResponse(&logFiles[0], env))
				expected, _ := json.Marshal(logResponses)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/logFiles?skip=1", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles", routing.GetAllLogFiles(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles", nil)
			q := req.URL.Query()
			q.Add("skip", "1")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 2 logFiles", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				logResponses := make([]*responses.LogFileResponse, 0)
				for _, model := range logFiles[1:] {
					logResponses = append(logResponses, responses.GenerateLogResponse(&model, env))
				}
				expected, _ := json.Marshal(logResponses)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/logFiles?limit=1", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles", routing.GetAllLogFiles(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles", nil)
			q := req.URL.Query()
			q.Add("limit", "1")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 2 logFiles", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				logResponses := make([]*responses.LogFileResponse, 0)
				logResponses = append(logResponses, responses.GenerateLogResponse(&logFiles[0], env))
				expected, _ := json.Marshal(logResponses)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/logFiles?orderBy=[timestamp]", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles", routing.GetAllLogFiles(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles", nil)
			q := req.URL.Query()
			q.Add("orderBy", "timestamp")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 3 logFiles", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				logResponses := make([]*responses.LogFileResponse, 0)
				for _, model := range logFiles {
					logResponses = append(logResponses, responses.GenerateLogResponse(&model, env))
				}
				expected, _ := json.Marshal(logResponses)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/logFiles?orderBy=[timestamp]&reverse=true", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles", routing.GetAllLogFiles(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles", nil)
			q := req.URL.Query()
			q.Add("orderBy", "timestamp")
			q.Add("reverse", "true")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 3 logFiles", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				logResponses := make([]*responses.LogFileResponse, 0)
				for _, model := range logFiles {
					logResponses = append(logResponses, responses.GenerateLogResponse(&model, env))
				}
				reversed := [3]*responses.LogFileResponse{
					logResponses[2],
					logResponses[1],
					logResponses[0],
				}
				expected, _ := json.Marshal(reversed)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})
	})
}

func TestCreateLogFile(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: dir}

		updateBody := &responses.LogFileCreationBody{}
		updateBody.Data.UUID = uuid.New().String()
		updateBody.Data.Name = "test"
		updateBody.Data.Note = "test note"

		env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
			return utils.IsValidUUID(fl.Field().String())
		})

		router := httprouter.New()
		Convey("Given a HTTP POST request for api/x/logFiles with a valid body", func() {
			router.POST("/api/x/logFiles", routing.CreateLogFile(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/logFiles", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a success status and a new logFile should have been created", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Logfiles"
				response.Body.Action = "CREATE"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/logFiles with an existing uuid", func() {
			updateBody.Data.UUID = logFiles[0].UUID
			router.POST("/api/x/logFiles", routing.CreateLogFile(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/logFiles", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail and return already exists error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Logfiles"
				response.Body.Action = "CREATE"
				response.Body.Error = fmt.Sprintf("Logfile with %v already exists", updateBody.Data.UUID)
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/logFiles with a invalid uuid in body", func() {
			router.POST("/api/x/logFiles", routing.CreateLogFile(env))
			updateBody.Data.UUID = "invalid"
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/logFiles", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with uuid validation fail error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Logfiles"
				response.Body.Action = "CREATE"
				response.Body.Error = "Key: 'LogFileCreationBody.Data.UUID' Error:Field validation for 'UUID' failed on the 'uuid' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/logFiles with a missing name in body", func() {
			router.POST("/api/x/logFiles", routing.CreateLogFile(env))
			updateBody.Data.Name = ""
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/logFiles", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with required name message", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Logfiles"
				response.Body.Action = "CREATE"
				response.Body.Error = "Key: 'LogFileCreationBody.Data.Name' Error:Field validation for 'Name' failed on the 'required' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestUpdateLogFile(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir)}

		env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
			return utils.IsValidUUID(fl.Field().String())
		})
		updateBody := &responses.LogFileCreationBody{}
		updateBody.Data.UUID = logFiles[0].UUID
		updateBody.Data.Name = logFiles[0].Name
		updateBody.Data.Note = "test note"

		router := httprouter.New()
		Convey("Given a HTTP PUT request for api/x/logFiles/uuid with a valid body", func() {
			router.PUT("/api/x/logFiles/:uuid", routing.UpdateLogFile(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/logFiles/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a successful", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Logfiles"
				response.Body.Action = "UPDATE"
				response.Body.Error = ""
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/logFiles/uuid with a unknown uid", func() {
			router.PUT("/api/x/logFiles/:uuid", routing.UpdateLogFile(env))
			updateBody.Data.UUID = "550e8400-e29b-41d4-a716-446655440004"
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/logFiles/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should return a not found error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusNotFound
				response.Body.Resource = "Logfiles"
				response.Body.Action = "UPDATE"
				response.Body.Error = "not found"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusNotFound)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/logFiles/uuid with a invalid uuid in the body", func() {
			router.PUT("/api/x/logFiles/:uuid", routing.UpdateLogFile(env))
			updateBody.Data.UUID = ""
			requestBody, _ := json.Marshal(updateBody.Data)
			updateBody.Data.UUID = logFiles[0].UUID
			req := httptest.NewRequest("PUT", "/api/x/logFiles/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should return a internal server error with a uuid validation error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Logfiles"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'LogFileCreationBody.Data.UUID' Error:Field validation for 'UUID' failed on the 'uuid' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/logFiles/uuid with a missing name", func() {
			router.PUT("/api/x/logFiles/:uuid", routing.UpdateLogFile(env))
			updateBody.Data.Name = ""

			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/logFiles/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with name validation error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Logfiles"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'LogFileCreationBody.Data.Name' Error:Field validation for 'Name' failed on the 'required' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestDeleteLogFile(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: dir}
		router := httprouter.New()
		Convey("Given a HTTP DELETE request for api/x/logFiles/uuid with a known uuid", func() {
			router.DELETE("/api/x/logFiles/:uuid", routing.DeleteLogFile(env))

			req := httptest.NewRequest("DELETE", "/api/x/logFiles/"+logFiles[0].UUID, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be successful", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Logfiles"
				response.Body.Action = "DELETE"
				response.Body.Error = ""
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP DELETE request for api/x/logFiles/uuid with a unknown uuid", func() {
			router.DELETE("/api/x/logFiles/:uuid", routing.DeleteLogFile(env))

			req := httptest.NewRequest("DELETE", "/api/x/logFiles/unknown", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail and return not found", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusNotFound
				response.Body.Resource = "Logfiles"
				response.Body.Action = "DELETE"
				response.Body.Error = "not found"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusNotFound)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}
