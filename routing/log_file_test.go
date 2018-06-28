package routing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/routing/responses"
	uuid "github.com/satori/go.uuid"

	"github.com/3devo/feconnector/utils"
	"github.com/asdine/storm"
	"github.com/julienschmidt/httprouter"
	. "github.com/smartystreets/goconvey/convey"
	validator "gopkg.in/go-playground/validator.v9"
)

var logFiles = []models.LogFile{
	{
		UUID:      "550e8400-e29b-41d4-a716-446655440000",
		Name:      "log1",
		Timestamp: 1,
		HasNote:   false},
	{
		UUID:      "550e8400-e29b-41d4-a716-446655440001",
		Name:      "log2",
		Timestamp: 2,
		HasNote:   false},
	{
		UUID:      "550e8400-e29b-41d4-a716-446655440002",
		Name:      "log3",
		Timestamp: 3,
		HasNote:   true}}

func TestGetSingle(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := prepareDB()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), FileDir: path.Dir(dir)}
		Convey("Given a HTTP request for /api/x/logFiles/550e8400-e29b-41d4-a716-446655440000", func() {

			router := httprouter.New()
			router.GET("/api/x/logFiles/:uuid", GetLogFile(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles/550e8400-e29b-41d4-a716-446655440000", nil)
			resp := httptest.NewRecorder()

			Convey("When the request is handled by the Router", func() {
				router.ServeHTTP(resp, req)

				Convey("Then the response should be a 200 because the item exists in the db", func() {
					result := resp.Result()
					body, _ := ioutil.ReadAll(result.Body)
					expected, _ := json.Marshal(responses.GenerateLogResponse(&logFiles[0], env))

					So(result.StatusCode, ShouldEqual, 200)
					So(body, ShouldResemble, append(expected, 10))
				})
			})
		})

		Convey("Given a HTTP request for /api/x/logFiles/undefined", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles/:uuid", GetLogFile(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles/undefined", nil)
			resp := httptest.NewRecorder()

			Convey("When the request is handled by the Router", func() {
				router.ServeHTTP(resp, req)

				Convey("Then the response should be a 404 with correct response body", func() {
					result := resp.Result()
					body, _ := ioutil.ReadAll(result.Body)
					response := responses.ResourceStatusResponse{}
					response.Body.Code = 404
					response.Body.Resource = "Logfiles"
					response.Body.Action = "GET"
					response.Body.Error = "not found"
					expected, _ := json.Marshal(response.Body)

					So(result.StatusCode, ShouldEqual, 404)
					So(string(body), ShouldResemble, string(append(expected, 10)))
				})
			})
		})
	})
}

func TestGetMultiple(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := prepareDB()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), FileDir: path.Dir(dir)}

		Convey("Given a HTTP request for api/x/logFiles", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles", GetAllLogFiles(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return 200 with the 3 logFiles", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				logResponses := make([]*responses.LogFileResponse, 0)
				for _, model := range logFiles {
					logResponses = append(logResponses, responses.GenerateLogResponse(&model, env))
				}
				expected, _ := json.Marshal(logResponses)

				So(result.StatusCode, ShouldEqual, 200)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/logFiles?filter=[{\"key\":\"name\", \"value\":\"log1\"}]", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles", GetAllLogFiles(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles", nil)
			q := req.URL.Query()
			q.Add("filter", "[{\"key\":\"name\", \"value\":\"log1\"}]")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return 200 with the 1 logFiles", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				logResponses := make([]*responses.LogFileResponse, 0)
				logResponses = append(logResponses, responses.GenerateLogResponse(&logFiles[0], env))
				expected, _ := json.Marshal(logResponses)

				So(result.StatusCode, ShouldEqual, 200)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/logFiles?skip=1", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles", GetAllLogFiles(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles", nil)
			q := req.URL.Query()
			q.Add("skip", "1")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return 200 with the 2 logFiles", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				logResponses := make([]*responses.LogFileResponse, 0)
				for _, model := range logFiles[1:] {
					logResponses = append(logResponses, responses.GenerateLogResponse(&model, env))
				}
				expected, _ := json.Marshal(logResponses)

				So(result.StatusCode, ShouldEqual, 200)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/logFiles?limit=1", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles", GetAllLogFiles(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles", nil)
			q := req.URL.Query()
			q.Add("limit", "1")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return 200 with the 2 logFiles", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				logResponses := make([]*responses.LogFileResponse, 0)
				logResponses = append(logResponses, responses.GenerateLogResponse(&logFiles[0], env))
				expected, _ := json.Marshal(logResponses)

				So(result.StatusCode, ShouldEqual, 200)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/logFiles?orderBy=[timestamp]", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles", GetAllLogFiles(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles", nil)
			q := req.URL.Query()
			q.Add("orderBy", "timestamp")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return 200 with the 3 logFiles", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				logResponses := make([]*responses.LogFileResponse, 0)
				for _, model := range logFiles {
					logResponses = append(logResponses, responses.GenerateLogResponse(&model, env))
				}
				expected, _ := json.Marshal(logResponses)

				So(result.StatusCode, ShouldEqual, 200)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/logFiles?orderBy=[timestamp]&reverse=true", func() {
			router := httprouter.New()
			router.GET("/api/x/logFiles", GetAllLogFiles(env))

			req := httptest.NewRequest("GET", "/api/x/logFiles", nil)
			q := req.URL.Query()
			q.Add("orderBy", "timestamp")
			q.Add("reverse", "true")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return 200 with the 3 logFiles", func() {
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

				So(result.StatusCode, ShouldEqual, 200)
				So(body, ShouldResemble, append(expected, 10))
			})
		})
	})
}

func TestCreate(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := prepareDB()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), FileDir: dir}

		updateBody := &responses.LogFileUpdateBody{}
		updateBody.Data.UUID = uuid.Must(uuid.NewV4()).String()
		updateBody.Data.Name = "test"
		updateBody.Data.Note = "test note"

		env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
			return utils.IsValidUUID(fl.Field().String())
		})

		router := httprouter.New()
		Convey("Given a HTTP POST request for api/x/logFiles with a valid body", func() {
			router.POST("/api/x/logFiles", CreateLogFile(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/logFiles", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a success status and a new logFile should have been created", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 200
				response.Body.Resource = "Logfiles"
				response.Body.Action = "CREATE"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 200)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/logFiles with an existing uuid", func() {
			updateBody.Data.UUID = logFiles[0].UUID
			router.POST("/api/x/logFiles", CreateLogFile(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/logFiles", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail and return already exists error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 500
				response.Body.Resource = "Logfiles"
				response.Body.Action = "CREATE"
				response.Body.Error = fmt.Sprintf("Logfile with %v already exists", updateBody.Data.UUID)
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 500)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/logFiles with a invalid uuid in body", func() {
			router.POST("/api/x/logFiles", CreateLogFile(env))
			updateBody.Data.UUID = "invalid"
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/logFiles", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with uuid validation fail error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 500
				response.Body.Resource = "Logfiles"
				response.Body.Action = "CREATE"
				response.Body.Error = "Key: 'LogFileUpdateBody.Data.UUID' Error:Field validation for 'UUID' failed on the 'uuid' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 500)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/logFiles with a missing name in body", func() {
			router.POST("/api/x/logFiles", CreateLogFile(env))
			updateBody.Data.Name = ""
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/logFiles", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with required name message", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 500
				response.Body.Resource = "Logfiles"
				response.Body.Action = "CREATE"
				response.Body.Error = "Key: 'LogFileUpdateBody.Data.Name' Error:Field validation for 'Name' failed on the 'required' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 500)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestUpdate(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := prepareDB()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), FileDir: path.Dir(dir)}

		env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
			return utils.IsValidUUID(fl.Field().String())
		})
		updateBody := &responses.LogFileUpdateBody{}
		updateBody.Data.UUID = logFiles[0].UUID
		updateBody.Data.Name = logFiles[0].Name
		updateBody.Data.Note = "test note"

		router := httprouter.New()
		Convey("Given a HTTP PUT request for api/x/logFiles/uuid with a valid body", func() {
			router.PUT("/api/x/logFiles/:uuid", UpdateLogFile(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/logFiles/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a successful", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 200
				response.Body.Resource = "Logfiles"
				response.Body.Action = "UPDATE"
				response.Body.Error = ""
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 200)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/logFiles/uuid with a unknown uid", func() {
			router.PUT("/api/x/logFiles/:uuid", UpdateLogFile(env))
			updateBody.Data.UUID = "550e8400-e29b-41d4-a716-446655440004"
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/logFiles/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should return a not found error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 404
				response.Body.Resource = "Logfiles"
				response.Body.Action = "UPDATE"
				response.Body.Error = "not found"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 404)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/logFiles/uuid with a invalid uuid in the body", func() {
			router.PUT("/api/x/logFiles/:uuid", UpdateLogFile(env))
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
				response.Body.Code = 500
				response.Body.Resource = "Logfiles"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'LogFileUpdateBody.Data.UUID' Error:Field validation for 'UUID' failed on the 'uuid' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 500)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/logFiles/uuid with a missing name", func() {
			router.PUT("/api/x/logFiles/:uuid", UpdateLogFile(env))
			updateBody.Data.Name = ""

			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/logFiles/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with name validation error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = 500
				response.Body.Resource = "Logfiles"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'LogFileUpdateBody.Data.Name' Error:Field validation for 'Name' failed on the 'required' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 500)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestDelete(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := prepareDB()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), FileDir: dir}
		router := httprouter.New()
		Convey("Given a HTTP DELETE request for api/x/logFiles/uuid with a known uuid", func() {
			router.DELETE("/api/x/logFiles/:uuid", DeleteLogFile(env))

			req := httptest.NewRequest("DELETE", "/api/x/logFiles/"+logFiles[0].UUID, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be successful", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = 200
				response.Body.Resource = "Logfiles"
				response.Body.Action = "DELETE"
				response.Body.Error = ""
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 200)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP DELETE request for api/x/logFiles/uuid with a unknown uuid", func() {
			router.DELETE("/api/x/logFiles/:uuid", DeleteLogFile(env))

			req := httptest.NewRequest("DELETE", "/api/x/logFiles/unknown", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail and return not found", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = 404
				response.Body.Resource = "Logfiles"
				response.Body.Action = "DELETE"
				response.Body.Error = "not found"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 404)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func prepareDB() (string, *storm.DB) {
	dir := filepath.Join(os.TempDir(), "feconnector-test")
	os.MkdirAll(dir, os.ModePerm)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	os.Mkdir(filepath.Join(dir, "logs"), os.ModePerm)
	os.Mkdir(filepath.Join(dir, "notes"), os.ModePerm)
	for _, model := range logFiles {
		db.Save(&model)
	}
	for _, model := range logFiles {
		logName := model.Name + "-" + time.Unix(model.Timestamp, 0).Format("2006-01-02-15-04-05") + ".txt"
		f, _ := os.Create(filepath.Join(dir, "logs", logName))
		f.Close()
		if model.HasNote {
			ioutil.WriteFile(filepath.Join(dir, "notes", logName), []byte("test"), os.ModePerm)
		}
	}
	return dir, db
}
