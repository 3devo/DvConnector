package routing

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/routing/responses"

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
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()
	env := &utils.Env{Db: db, Validator: validator.New()}

	Convey("Given a HTTP request for /api/x/logFiles/550e8400-e29b-41d4-a716-446655440000", t, func() {

		router := httprouter.New()
		router.GET("/api/x/logFiles/:uuid", GetLogFile(env))

		req := httptest.NewRequest("GET", "/api/x/logFiles/550e8400-e29b-41d4-a716-446655440000", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			router.ServeHTTP(resp, req)

			Convey("Then the response should be a 200 because the item exists in the db", func() {
				result := resp.Result()
				So(result.StatusCode, ShouldEqual, 200)
				body, _ := ioutil.ReadAll(result.Body)
				expected, _ := json.Marshal(responses.GenerateLogResponse(&logFiles[0], env))
				So(body, ShouldResemble, append(expected, 10))
			})
		})
	})

	Convey("Given a HTTP request for /api/x/logFiles/undefined", t, func() {
		router := httprouter.New()
		router.GET("/api/x/logFiles/:uuid", GetLogFile(env))

		req := httptest.NewRequest("GET", "/api/x/logFiles/undefined", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			router.ServeHTTP(resp, req)

			Convey("Then the response should be a 404 with correct response body", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				So(result.StatusCode, ShouldEqual, 404)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 404
				response.Body.Resource = "Logfiles"
				response.Body.Action = "GET"
				response.Body.Error = "not found"
				expected, _ := json.Marshal(response.Body)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestGetMultiple(t *testing.T) {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()
	env := &utils.Env{Db: db, Validator: validator.New()}

	Convey("Given a HTTP request for api/x/logFiles", t, func() {
		router := httprouter.New()
		router.GET("/api/x/logFiles", GetAllLogFiles(env))

		req := httptest.NewRequest("GET", "/api/x/logFiles", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		Convey("Then the response should return 200 with the 3 logFiles", func() {
			result := resp.Result()
			So(result.StatusCode, ShouldEqual, 200)
			body, _ := ioutil.ReadAll(result.Body)
			logResponses := make([]*responses.LogFileResponse, 0)
			for _, model := range logFiles {
				logResponses = append(logResponses, responses.GenerateLogResponse(&model, env))
			}
			expected, _ := json.Marshal(logResponses)
			So(body, ShouldResemble, append(expected, 10))
		})
	})

	Convey("Given a HTTP request for api/x/logFiles?filter=[{\"key\":\"name\", \"value\":\"log1\"}]", t, func() {
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
			So(result.StatusCode, ShouldEqual, 200)
			body, _ := ioutil.ReadAll(result.Body)
			logResponses := make([]*responses.LogFileResponse, 0)
			logResponses = append(logResponses, responses.GenerateLogResponse(&logFiles[0], env))
			expected, _ := json.Marshal(logResponses)
			So(body, ShouldResemble, append(expected, 10))
		})
	})

	Convey("Given a HTTP request for api/x/logFiles?skip=1", t, func() {
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
			So(result.StatusCode, ShouldEqual, 200)
			body, _ := ioutil.ReadAll(result.Body)
			logResponses := make([]*responses.LogFileResponse, 0)
			for _, model := range logFiles[1:] {
				logResponses = append(logResponses, responses.GenerateLogResponse(&model, env))
			}
			expected, _ := json.Marshal(logResponses)
			So(body, ShouldResemble, append(expected, 10))
		})
	})

	Convey("Given a HTTP request for api/x/logFiles?limit=1", t, func() {
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
			So(result.StatusCode, ShouldEqual, 200)
			body, _ := ioutil.ReadAll(result.Body)
			logResponses := make([]*responses.LogFileResponse, 0)
			logResponses = append(logResponses, responses.GenerateLogResponse(&logFiles[0], env))
			expected, _ := json.Marshal(logResponses)
			So(body, ShouldResemble, append(expected, 10))
		})
	})

	Convey("Given a HTTP request for api/x/logFiles?orderBy=[timestamp]", t, func() {
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
			So(result.StatusCode, ShouldEqual, 200)
			body, _ := ioutil.ReadAll(result.Body)
			logResponses := make([]*responses.LogFileResponse, 0)
			for _, model := range logFiles {
				logResponses = append(logResponses, responses.GenerateLogResponse(&model, env))
			}
			expected, _ := json.Marshal(logResponses)
			So(body, ShouldResemble, append(expected, 10))
		})
	})

	Convey("Given a HTTP request for api/x/logFiles?orderBy=[timestamp]&reverse=true", t, func() {
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
			So(result.StatusCode, ShouldEqual, 200)
			body, _ := ioutil.ReadAll(result.Body)
			logResponses := make([]*responses.LogFileResponse, 0)
			for _, model := range logFiles {
				logResponses = append(logResponses, responses.GenerateLogResponse(&model, env))
			}
			for i := len(logResponses)/2 - 1; i >= 0; i-- {
				opp := len(logResponses) - 1 - i
				logResponses[i], logResponses[opp] = logResponses[opp], logResponses[i]
			}
			expected, _ := json.Marshal(logResponses)
			So(body, ShouldResemble, append(expected, 10))
		})
	})
}

func TestCreate(t *testing.T) {
	// dir, db := prepareDB()
	// defer os.RemoveAll(dir)
	// defer db.Close()
	// env := &utils.Env{Db: db, Validator: validator.New()}

	// Convey("Given a HTTP POST request for api/x/logFiles", t, func() {
	// 	env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
	// 		return utils.IsValidUUID(fl.Field().String())
	// 	})

	// 	router := httprouter.New()
	// 	router.POST("/api/x/logFiles", CreateLogFile(env))
	// 	updateBody := &responses.LogFileUpdateBody{}
	// 	updateBody.Data.UUID = uuid.Must(uuid.NewV4()).String()
	// 	updateBody.Data.Name = "test"
	// 	updateBody.Data.Note = "test note"
	// 	requestBody, _ := json.Marshal(updateBody.Data)
	// 	req := httptest.NewRequest("POST", "/api/x/logFiles", strings.NewReader(string(requestBody)))
	// 	resp := httptest.NewRecorder()

	// 	router.ServeHTTP(resp, req)

	// 	Convey("Then the response should be a success status and a new logFile should have been created", func() {
	// 		result := resp.Result()
	// 		body, _ := ioutil.ReadAll(result.Body)
	// 		log.Println(string(requestBody), string(body))
	// 		So(result.StatusCode, ShouldEqual, 200)
	// 		response := responses.ResourceStatusResponse{}
	// 		response.Body.Code = 200
	// 		response.Body.Resource = "Logfiles"
	// 		response.Body.Action = "CREATE"
	// 		expected, _ := json.Marshal(response.Body)
	// 		So(string(body), ShouldResemble, string(append(expected, 10)))
	// 	})
	// })
}

func prepareDB() (string, *storm.DB) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	for _, model := range logFiles {
		db.Save(&model)
	}
	return dir, db
}
