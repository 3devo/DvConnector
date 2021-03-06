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

func TestGetSingleChart(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir)}
		Convey("Given a HTTP request for /api/x/charts/550e8400-e29b-41d4-a716-446655440000", func() {

			router := httprouter.New()
			router.GET("/api/x/charts/:uuid", routing.GetChart(env))

			req := httptest.NewRequest("GET", "/api/x/charts/550e8400-e29b-41d4-a716-446655440000", nil)
			resp := httptest.NewRecorder()

			Convey("When the request is handled by the Router", func() {
				router.ServeHTTP(resp, req)

				Convey("Then the response should be a http.StatusOK because the item exists in the db", func() {
					result := resp.Result()
					body, _ := ioutil.ReadAll(result.Body)
					expected, _ := json.Marshal(charts[0])
					So(result.StatusCode, ShouldEqual, http.StatusOK)
					So(body, ShouldResemble, append(expected, 10))
				})
			})
		})

		Convey("Given a HTTP request for /api/x/charts/undefined", func() {
			router := httprouter.New()
			router.GET("/api/x/charts/:uuid", routing.GetChart(env))

			req := httptest.NewRequest("GET", "/api/x/charts/undefined", nil)
			resp := httptest.NewRecorder()

			Convey("When the request is handled by the Router", func() {
				router.ServeHTTP(resp, req)

				Convey("Then the response should be a http.StatusNotFound with correct response body", func() {
					result := resp.Result()
					body, _ := ioutil.ReadAll(result.Body)
					response := responses.ResourceStatusResponse{}
					response.Body.Code = http.StatusNotFound
					response.Body.Resource = "Charts"
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

func TestGetMultipleCharts(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir)}

		Convey("Given a HTTP request for api/x/charts", func() {
			router := httprouter.New()
			router.GET("/api/x/charts", routing.GetAllCharts(env))

			req := httptest.NewRequest("GET", "/api/x/charts", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 3 charts", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				expected, _ := json.Marshal(&charts)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/charts?filter=[{\"key\":\"title\", \"value\":\"chart0\"}]", func() {
			router := httprouter.New()
			router.GET("/api/x/charts", routing.GetAllCharts(env))

			req := httptest.NewRequest("GET", "/api/x/charts", nil)
			q := req.URL.Query()
			q.Add("filter", "[{\"key\":\"title\", \"value\":\"chart0\"}]")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 1 charts", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				responseBody := []*models.Chart{&charts[0]}
				expected, _ := json.Marshal(responseBody)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/charts?skip=1", func() {
			router := httprouter.New()
			router.GET("/api/x/charts", routing.GetAllCharts(env))

			req := httptest.NewRequest("GET", "/api/x/charts", nil)
			q := req.URL.Query()
			q.Add("skip", "1")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 2 charts", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				var responseBody = []*models.Chart{
					&charts[1],
					&charts[2]}
				expected, _ := json.Marshal(responseBody)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP request for api/x/charts?limit=1", func() {
			router := httprouter.New()
			router.GET("/api/x/charts", routing.GetAllCharts(env))

			req := httptest.NewRequest("GET", "/api/x/charts", nil)
			q := req.URL.Query()
			q.Add("limit", "1")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 2 charts", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				var responseBody = []*models.Chart{
					&charts[0]}
				expected, _ := json.Marshal(responseBody)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/charts?orderBy=[title]", func() {
			router := httprouter.New()
			router.GET("/api/x/charts", routing.GetAllCharts(env))

			req := httptest.NewRequest("GET", "/api/x/charts", nil)
			q := req.URL.Query()
			q.Add("orderBy", "title")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 3 charts", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				expected, _ := json.Marshal(charts)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})

		Convey("Given a HTTP request for api/x/charts?orderBy=[title]&reverse=true", func() {
			router := httprouter.New()
			router.GET("/api/x/charts", routing.GetAllCharts(env))

			req := httptest.NewRequest("GET", "/api/x/charts", nil)
			q := req.URL.Query()
			q.Add("orderBy", "title")
			q.Add("reverse", "true")
			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Convey("Then the response should return http.StatusOK with the 3 charts", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				reversed := [3]*models.Chart{
					&charts[2],
					&charts[1],
					&charts[0],
				}
				expected, _ := json.Marshal(reversed)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(body, ShouldResemble, append(expected, 10))
			})
		})
	})
}

func TestCreateChart(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: dir}

		updateBody := &responses.ChartCreationBody{Data: CreateChart(&charts[0])}
		updateBody.Data.UUID = "550e8400-e29b-41d4-a716-446655440003"

		env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
			return utils.IsValidUUID(fl.Field().String())
		})

		router := httprouter.New()
		Convey("Given a HTTP POST request for api/x/charts with a valid body", func() {
			router.POST("/api/x/charts", routing.CreateChart(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/charts", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a success status and a new chart should have been created", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Charts"
				response.Body.Action = "CREATE"
				expected, _ := json.Marshal(response.Body)

				So(string(body), ShouldResemble, string(append(expected, 10)))
				So(result.StatusCode, ShouldEqual, http.StatusOK)
			})
		})

		Convey("Given a HTTP POST request for api/x/charts with an existing uuid", func() {
			updateBody.Data.UUID = charts[0].UUID
			router.POST("/api/x/charts", routing.CreateChart(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/charts", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail and return already exists error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Charts"
				response.Body.Action = "CREATE"
				response.Body.Error = fmt.Sprintf("Chart with %v already exists", updateBody.Data.UUID)
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/charts with a invalid uuid in body", func() {
			router.POST("/api/x/charts", routing.CreateChart(env))
			updateBody.Data.UUID = "invalid"
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/charts", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with uuid validation fail error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Charts"
				response.Body.Action = "CREATE"
				response.Body.Error = "Key: 'ChartCreationBody.Data.UUID' Error:Field validation for 'UUID' failed on the 'uuid' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP POST request for api/x/charts with a missing name in body", func() {
			router.POST("/api/x/charts", routing.CreateChart(env))
			updateBody.Data.Title = ""
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("POST", "/api/x/charts", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with required name message", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Charts"
				response.Body.Action = "CREATE"
				response.Body.Error = "Key: 'ChartCreationBody.Data.Title' Error:Field validation for 'Title' failed on the 'required' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestUpdateChart(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir)}

		env.Validator.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
			return utils.IsValidUUID(fl.Field().String())
		})
		updateBody := responses.ChartCreationBody{Data: CreateChart(&charts[0])}

		router := httprouter.New()
		Convey("Given a HTTP PUT request for api/x/charts/uuid with a valid body", func() {
			router.PUT("/api/x/charts/:uuid", routing.UpdateChart(env))
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/charts/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a successful", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Charts"
				response.Body.Action = "UPDATE"
				response.Body.Error = ""
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/charts/uuid with a unknown uid", func() {
			router.PUT("/api/x/charts/:uuid", routing.UpdateChart(env))
			updateBody.Data.UUID = "550e8400-e29b-41d4-a716-446655440004"
			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/charts/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should return a not found error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusNotFound
				response.Body.Resource = "Charts"
				response.Body.Action = "UPDATE"
				response.Body.Error = "not found"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusNotFound)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP PUT request for api/x/charts/uuid with a invalid uuid in the body", func() {
			router.PUT("/api/x/charts/:uuid", routing.UpdateChart(env))
			updateBody.Data.UUID = ""
			requestBody, _ := json.Marshal(updateBody.Data)
			updateBody.Data.UUID = charts[0].UUID
			req := httptest.NewRequest("PUT", "/api/x/charts/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should return a internal server error with a uuid validation error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Charts"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'ChartCreationBody.Data.UUID' Error:Field validation for 'UUID' failed on the 'uuid' tag"
				expected, _ := json.Marshal(response.Body)

				So(string(body), ShouldResemble, string(append(expected, 10)))
				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Given a HTTP PUT request for api/x/charts/uuid with a missing name", func() {
			router.PUT("/api/x/charts/:uuid", routing.UpdateChart(env))
			updateBody.Data.Title = ""

			requestBody, _ := json.Marshal(updateBody.Data)
			req := httptest.NewRequest("PUT", "/api/x/charts/"+updateBody.Data.UUID, strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be a internal server error with name validation error", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Charts"
				response.Body.Action = "UPDATE"
				response.Body.Error = "Key: 'ChartCreationBody.Data.Title' Error:Field validation for 'Title' failed on the 'required' tag"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestDeleteChart(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: dir}
		router := httprouter.New()
		Convey("Given a HTTP DELETE request for api/x/charts/uuid with a known uuid", func() {
			router.DELETE("/api/x/charts/:uuid", routing.DeleteChart(env))

			req := httptest.NewRequest("DELETE", "/api/x/charts/"+charts[0].UUID, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should be successful", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Charts"
				response.Body.Action = "DELETE"
				response.Body.Error = ""
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP DELETE request for api/x/charts/uuid with a unknown uuid", func() {
			router.DELETE("/api/x/charts/:uuid", routing.DeleteChart(env))

			req := httptest.NewRequest("DELETE", "/api/x/charts/unknown", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail and return not found", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusNotFound
				response.Body.Resource = "Charts"
				response.Body.Action = "DELETE"
				response.Body.Error = "not found"
				expected, _ := json.Marshal(response.Body)

				So(result.StatusCode, ShouldEqual, http.StatusNotFound)
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}
