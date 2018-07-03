package routing_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/3devo/feconnector/routing"
	"github.com/3devo/feconnector/routing/responses"

	"github.com/3devo/feconnector/utils"
	"github.com/julienschmidt/httprouter"
	. "github.com/smartystreets/goconvey/convey"
	validator "gopkg.in/go-playground/validator.v9"
)

func TestGetSingleSheet(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), FileDir: path.Dir(dir)}
		Convey("Given a HTTP request for /api/x/sheets/550e8400-e29b-41d4-a716-446655440000", func() {

			router := httprouter.New()
			router.GET("/api/x/sheets/:uuid", routing.GetSheet(env))

			req := httptest.NewRequest("GET", "/api/x/sheets/550e8400-e29b-41d4-a716-446655440000", nil)
			resp := httptest.NewRecorder()

			Convey("When the request is handled by the Router", func() {
				router.ServeHTTP(resp, req)

				Convey("Then the response should be a 200 because the item exists in the db", func() {
					result := resp.Result()
					body, _ := ioutil.ReadAll(result.Body)
					expected, _ := json.Marshal(responses.GenerateSheetResponseObject(&sheets[0], env))
					So(result.StatusCode, ShouldEqual, 200)
					So(body, ShouldResemble, append(expected, 10))
				})
			})
		})

		Convey("Given a HTTP request for /api/x/sheets/undefined", func() {
			router := httprouter.New()
			router.GET("/api/x/sheets/:uuid", routing.GetSheet(env))

			req := httptest.NewRequest("GET", "/api/x/sheets/undefined", nil)
			resp := httptest.NewRecorder()

			Convey("When the request is handled by the Router", func() {
				router.ServeHTTP(resp, req)

				Convey("Then the response should be a 404 with correct response body", func() {
					result := resp.Result()
					body, _ := ioutil.ReadAll(result.Body)
					response := responses.ResourceStatusResponse{}
					response.Body.Code = 404
					response.Body.Resource = "Sheets"
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

func TestGetMultipleSheets(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), FileDir: path.Dir(dir)}

		Convey("Given a HTTP request for api/x/sheets", func() {
			router := httprouter.New()
			router.GET("/api/x/sheets", routing.GetAllSheets(env))

			req := httptest.NewRequest("GET", "/api/x/sheets", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return 200 with the 3 sheets", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				responseBody := make([]*responses.SheetResponse, 0)
				for _, model := range sheets {
					responseBody = append(responseBody, responses.GenerateSheetResponseObject(&model, env))
				}
				expected, _ := json.Marshal(responseBody)
				So(result.StatusCode, ShouldEqual, 200)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/sheets?filter=[{\"key\":\"title\", \"value\":\"sheet0\"}]", func() {
			router := httprouter.New()
			router.GET("/api/x/sheets", routing.GetAllSheets(env))

			req := httptest.NewRequest("GET", "/api/x/sheets", nil)
			q := req.URL.Query()
			q.Add("filter", "[{\"key\":\"title\", \"value\":\"sheet0\"}]")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return 200 with the 1 sheets", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				expected, _ := json.Marshal([]*responses.SheetResponse{responses.GenerateSheetResponseObject(&sheets[0], env)})

				So(result.StatusCode, ShouldEqual, 200)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/sheets?skip=1", func() {
			router := httprouter.New()
			router.GET("/api/x/sheets", routing.GetAllSheets(env))

			req := httptest.NewRequest("GET", "/api/x/sheets", nil)
			q := req.URL.Query()
			q.Add("skip", "1")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return 200 with the 2 sheets", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				responseBody := make([]*responses.SheetResponse, 0)
				for _, model := range sheets[1:] {
					responseBody = append(responseBody, responses.GenerateSheetResponseObject(&model, env))
				}
				expected, _ := json.Marshal(responseBody)

				So(result.StatusCode, ShouldEqual, 200)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP request for api/x/sheets?limit=1", func() {
			router := httprouter.New()
			router.GET("/api/x/sheets", routing.GetAllSheets(env))

			req := httptest.NewRequest("GET", "/api/x/sheets", nil)
			q := req.URL.Query()
			q.Add("limit", "1")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return 200 with the 2 sheets", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				expected, _ := json.Marshal([]*responses.SheetResponse{responses.GenerateSheetResponseObject(&sheets[0], env)})

				So(result.StatusCode, ShouldEqual, 200)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/sheets?orderBy=[title]", func() {
			router := httprouter.New()
			router.GET("/api/x/sheets", routing.GetAllSheets(env))

			req := httptest.NewRequest("GET", "/api/x/sheets", nil)
			q := req.URL.Query()
			q.Add("orderBy", "title")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return 200 with the 3 sheets", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				responseBody := make([]*responses.SheetResponse, 0)
				for _, model := range sheets {
					responseBody = append(responseBody, responses.GenerateSheetResponseObject(&model, env))
				}
				expected, _ := json.Marshal(responseBody)
				So(result.StatusCode, ShouldEqual, 200)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/sheets?orderBy=[title]&reverse=true", func() {
			router := httprouter.New()
			router.GET("/api/x/sheets", routing.GetAllSheets(env))

			req := httptest.NewRequest("GET", "/api/x/sheets", nil)
			q := req.URL.Query()
			q.Add("orderBy", "title")
			q.Add("reverse", "true")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return 200 with the 3 sheets", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				responseBody := make([]*responses.SheetResponse, 0)
				for _, model := range sheets {
					responseBody = append(responseBody, responses.GenerateSheetResponseObject(&model, env))
				}
				reversed := [3]*responses.SheetResponse{
					responseBody[2],
					responseBody[1],
					responseBody[0],
				}
				expected, _ := json.Marshal(reversed)
				So(result.StatusCode, ShouldEqual, 200)
				So(body, ShouldResemble, append(expected, 10))
			})
		})
	})
}

func TestCreateSheet(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), FileDir: dir}

		updateBody := &responses.SheetCreationBody{Data: sheets[0]}
		updateBody.Data.UUID = "550e8400-e29b-41d4-a716-446655440003"
		env.Validator.RegisterValidation("chart-exists", func(fl validator.FieldLevel) bool {
			return true
		})
		router := httprouter.New()
		Convey("Given a HTTP POST request for api/x/sheets with a valid body", func() {
			env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
				return true
			})
			router.POST("/api/x/sheets", routing.CreateSheet(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/sheets", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a success status and a new sheet should have been created", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 200
				response.Body.Resource = "Sheets"
				response.Body.Action = "CREATE"
				expected, _ := json.Marshal(response.Body)

				So(string(body), ShouldResemble, string(append(expected, 10)))
				So(result.StatusCode, ShouldEqual, 200)
			})
		})

		Convey("Given a HTTP POST request for api/x/sheets with an existing uuid", func() {
			env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
				return true
			})
			updateBody.Data.UUID = sheets[0].UUID
			router.POST("/api/x/sheets", routing.CreateSheet(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/sheets", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail and return already exists error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 500
				response.Body.Resource = "Sheets"
				response.Body.Action = "CREATE"
				response.Body.Error = fmt.Sprintf("Sheet with %v already exists", updateBody.Data.UUID)
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 500)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/sheets with a invalid uuid in body", func() {
			env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
				return false
			})
			router.POST("/api/x/sheets", routing.CreateSheet(env))
			updateBody.Data.UUID = "invalid"
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/sheets", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with uuid validation fail error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 500
				response.Body.Resource = "Sheets"
				response.Body.Action = "CREATE"
				response.Body.Error = "Key: 'SheetCreationBody.Data.UUID' Error:Field validation for 'UUID' failed on the 'uuid' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 500)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/sheets with a missing name in body", func() {
			router.POST("/api/x/sheets", routing.CreateSheet(env))
			updateBody.Data.Title = ""
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/sheets", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with required name message", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 500
				response.Body.Resource = "Sheets"
				response.Body.Action = "CREATE"
				response.Body.Error = "Key: 'SheetCreationBody.Data.Title' Error:Field validation for 'Title' failed on the 'required' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 500)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/sheets with non existant chart", func() {
			env.Validator.RegisterValidation("chart-exists", func(fl validator.FieldLevel) bool {
				return false
			})
			router.POST("/api/x/sheets", routing.CreateSheet(env))
			updateBody.Data.Charts = []string{"unknown"}
			requestBody, _ := json.Marshal(updateBody.Data)

			req := httptest.NewRequest("POST", "/api/x/sheets", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with required name message", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 500
				response.Body.Resource = "Sheets"
				response.Body.Action = "CREATE"
				response.Body.Error = "Key: 'SheetCreationBody.Data.Charts[0]' Error:Field validation for 'Charts[0]' failed on the 'chart-exists' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 500)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestUpdateSheet(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), FileDir: path.Dir(dir)}

		env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
			return utils.IsValidUUID(fl.Field().String())
		})
		env.Validator.RegisterValidation("chart-exists", func(fl validator.FieldLevel) bool {
			return true
		})
		updateBody := responses.SheetCreationBody{Data: sheets[0]}

		router := httprouter.New()
		Convey("Given a HTTP PUT request for api/x/sheets/uuid with a valid body", func() {
			router.PUT("/api/x/sheets/:uuid", routing.UpdateSheet(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/sheets/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a successful", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 200
				response.Body.Resource = "Sheets"
				response.Body.Action = "UPDATE"
				response.Body.Error = ""
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 200)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/sheets/uuid with a unknown uid", func() {
			router.PUT("/api/x/sheets/:uuid", routing.UpdateSheet(env))
			updateBody.Data.UUID = "550e8400-e29b-41d4-a716-446655440004"
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/sheets/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should return a not found error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = 404
				response.Body.Resource = "Sheets"
				response.Body.Action = "UPDATE"
				response.Body.Error = "not found"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 404)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/sheets/uuid with a invalid uuid in the body", func() {
			router.PUT("/api/x/sheets/:uuid", routing.UpdateSheet(env))
			updateBody.Data.UUID = ""
			requestBody, _ := json.Marshal(updateBody.Data)
			updateBody.Data.UUID = sheets[0].UUID
			req := httptest.NewRequest("PUT", "/api/x/sheets/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should return a internal server error with a uuid validation error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = 500
				response.Body.Resource = "Sheets"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'SheetCreationBody.Data.UUID' Error:Field validation for 'UUID' failed on the 'uuid' tag"
				expected, _ := json.Marshal(response.Body)

				So(string(body), ShouldResemble, string(append(expected, 10)))
				So(result.StatusCode, ShouldEqual, 500)
			})
		})

		Convey("Given a HTTP PUT request for api/x/sheets/uuid with a missing name", func() {
			router.PUT("/api/x/sheets/:uuid", routing.UpdateSheet(env))
			updateBody.Data.Title = ""

			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/sheets/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with name validation error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = 500
				response.Body.Resource = "Sheets"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'SheetCreationBody.Data.Title' Error:Field validation for 'Title' failed on the 'required' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 500)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/sheets/uuid with a unknown chart", func() {
			env.Validator.RegisterValidation("chart-exists", func(fl validator.FieldLevel) bool {
				return false
			})

			router.PUT("/api/x/sheets/:uuid", routing.UpdateSheet(env))

			updateBody.Data.Charts = []string{"unknown"}
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/sheets/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with name validation error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = 500
				response.Body.Resource = "Sheets"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'SheetCreationBody.Data.Charts[0]' Error:Field validation for 'Charts[0]' failed on the 'chart-exists' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 500)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestDeleteSheet(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), FileDir: dir}
		router := httprouter.New()
		Convey("Given a HTTP DELETE request for api/x/sheets/uuid with a known uuid", func() {
			router.DELETE("/api/x/sheets/:uuid", routing.DeleteSheet(env))

			req := httptest.NewRequest("DELETE", "/api/x/sheets/"+sheets[0].UUID, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be successful", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = 200
				response.Body.Resource = "Sheets"
				response.Body.Action = "DELETE"
				response.Body.Error = ""
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 200)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP DELETE request for api/x/sheets/uuid with a unknown uuid", func() {
			router.DELETE("/api/x/sheets/:uuid", routing.DeleteSheet(env))

			req := httptest.NewRequest("DELETE", "/api/x/sheets/unknown", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail and return not found", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = 404
				response.Body.Resource = "Sheets"
				response.Body.Action = "DELETE"
				response.Body.Error = "not found"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, 404)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}
