package routing_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/3devo/feconnector/middleware"

	"github.com/tidwall/gjson"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/routing/responses"
	"github.com/google/uuid"

	"github.com/3devo/feconnector/routing"

	"github.com/julienschmidt/httprouter"

	"github.com/3devo/feconnector/utils"
	. "github.com/smartystreets/goconvey/convey"
	validator "gopkg.in/go-playground/validator.v9"
)

func TestLogin(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir)}

		Convey("Given a HTTP request for /api/x/login with the incorrect user information", func() {
			router := httprouter.New()
			router.POST("/api/x/login", routing.Login(env))
			loginParameters := responses.LoginParameters{
				Username: "bob",
				Password: "password"}
			requestBody, _ := json.Marshal(loginParameters)
			req := httptest.NewRequest("POST", "/api/x/login", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail because there is no user", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusUnauthorized
				response.Body.Resource = "Authentication"
				response.Body.Action = "LOGIN"
				response.Body.Error = "Invalid password"
				expected, _ := json.Marshal(response.Body)

				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP request for /api/x/login without a password", func() {
			router := httprouter.New()
			router.POST("/api/x/login", routing.Login(env))
			loginParameters := responses.LoginParameters{
				Username: "bob"}
			requestBody, _ := json.Marshal(loginParameters)
			req := httptest.NewRequest("POST", "/api/x/login", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail with a validation error on password required", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusInternalServerError
				response.Body.Resource = "Authentication"
				response.Body.Action = "LOGIN"
				response.Body.Error = "Key: 'LoginParameters.Password' Error:Field validation for 'Password' failed on the 'required' tag"
				expected, _ := json.Marshal(response.Body)

				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP request for /api/x/login with the correct user information", func() {
			passwordHash := utils.HashPassword("password")
			user := models.User{
				UUID:     uuid.New().String(),
				Username: "bob",
				Password: passwordHash}
			db.Save(&user)

			router := httprouter.New()
			router.POST("/api/x/login", routing.Login(env))
			loginParameters := responses.LoginParameters{
				Username: "bob",
				Password: "password"}
			requestBody, _ := json.Marshal(loginParameters)
			req := httptest.NewRequest("POST", "/api/x/login", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should succeed and return token that expires in an hour", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				tokenString := gjson.ParseBytes(body).Get("token").String()
				So(tokenString, ShouldNotBeEmpty)
				token, err := utils.ValidateJWTToken(tokenString)
				So(err, ShouldBeNil)
				So(token.ExpiresAt, ShouldAlmostEqual, time.Now().Add(time.Minute*time.Duration(utils.StandardTokenExpiration)).Unix())
			})
		})

		Convey("Given a HTTP request for /api/x/login with the correct user information and remember me", func() {
			passwordHash := utils.HashPassword("password")
			user := models.User{
				UUID:     uuid.New().String(),
				Username: "bob",
				Password: passwordHash}
			db.Save(&user)

			router := httprouter.New()
			router.POST("/api/x/login", routing.Login(env))
			loginParameters := responses.LoginParameters{
				Username:   "bob",
				Password:   "password",
				RememberMe: true}
			requestBody, _ := json.Marshal(loginParameters)
			req := httptest.NewRequest("POST", "/api/x/login", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should succeed and return token that expires in a month", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				tokenString := gjson.ParseBytes(body).Get("token").String()
				So(tokenString, ShouldNotBeEmpty)
				token, err := utils.ValidateJWTToken(tokenString)
				So(err, ShouldBeNil)
				So(token.ExpiresAt, ShouldAlmostEqual, time.Now().Add(time.Hour*time.Duration(utils.ExtendedTokenExpiration)).Unix())
			})
		})

		Convey("Given a HTTP request for /api/x/login with the invalid password", func() {
			passwordHash := utils.HashPassword("password")
			user := models.User{
				UUID:     uuid.New().String(),
				Username: "bob",
				Password: passwordHash}
			db.Save(&user)

			router := httprouter.New()
			router.POST("/api/x/login", routing.Login(env))
			loginParameters := responses.LoginParameters{
				Username: "bob",
				Password: "wrong"}
			requestBody, _ := json.Marshal(loginParameters)
			req := httptest.NewRequest("POST", "/api/x/login", strings.NewReader(string(requestBody)))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Then the response should fail because the password is invalid", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusUnauthorized
				response.Body.Resource = "Authentication"
				response.Body.Action = "LOGIN"
				response.Body.Error = "Invalid password"
				expected, _ := json.Marshal(response.Body)

				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestLogout(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir), HasAuth: true}

		Convey("Given a HTTP request for /api/x/logout without a authentication token and auth enabled", func() {
			router := httprouter.New()
			router.POST("/api/x/logout", middleware.AuthRequired(routing.Logout(env), env))
			req := httptest.NewRequest("POST", "/api/x/logout", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("The response should return unauthorized", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)

				So(string(body), ShouldResemble, "Unauthorized\n")
			})
		})

		Convey("Given a HTTP request for /api/x/logout with token but no auth required", func() {
			env.HasAuth = false
			router := httprouter.New()
			router.POST("/api/x/logout", middleware.AuthRequired(routing.Logout(env), env))
			req := httptest.NewRequest("POST", "/api/x/logout", nil)
			token, _ := utils.GenerateJWTToken("uuid", time.Now().Add(time.Minute*time.Duration(utils.StandardTokenExpiration)).Unix())
			req.Header.Add("Authorization", "bearer "+token)

			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("The response should return OK and do nothing", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Authentication"
				response.Body.Action = "LOGOUT"
				expected, _ := json.Marshal(response.Body)
				So(db.One("Token", token, &models.BlackListedToken{}), ShouldResemble, errors.New("not found"))
				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP request for /api/x/logout with a authentication token and auth enabled", func() {
			router := httprouter.New()
			router.POST("/api/x/logout", middleware.AuthRequired(routing.Logout(env), env))
			req := httptest.NewRequest("POST", "/api/x/logout", nil)
			token, _ := utils.GenerateJWTToken("uuid", time.Now().Add(time.Minute*time.Duration(utils.StandardTokenExpiration)).Unix())
			req.Header.Add("Authorization", "bearer "+token)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("The response should return OK and add the token to the blacklist", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.ResourceStatusResponse{}
				response.Body.Code = http.StatusOK
				response.Body.Resource = "Authentication"
				response.Body.Action = "LOGOUT"
				expected, _ := json.Marshal(response.Body)

				So(string(body), ShouldResemble, string(append(expected, 10)))
				So(db.One("Token", token, &models.BlackListedToken{}), ShouldBeNil)
			})
		})
	})
}

func TestAuthRequired(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir), HasAuth: false}
		Convey("Given a HTTP request for /api/x/authRequired with auth disabled", func() {
			router := httprouter.New()
			router.GET("/api/x/authRequired", routing.AuthRequired(env))
			req := httptest.NewRequest("GET", "/api/x/authRequired", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("The response should return enabled false and return OK", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.AuthEnabledResponse{Enabled: false}

				expected, _ := json.Marshal(response)

				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})

		Convey("Given a HTTP request for /api/x/authRequired with auth enabled", func() {
			env.HasAuth = true
			router := httprouter.New()
			router.GET("/api/x/authRequired", routing.AuthRequired(env))
			req := httptest.NewRequest("GET", "/api/x/authRequired", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("The response should return enabled false and return OK", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				response := responses.AuthEnabledResponse{Enabled: true}

				expected, _ := json.Marshal(response)

				So(string(body), ShouldResemble, string(append(expected, 10)))
			})
		})
	})
}

func TestRefreshToken(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		defer os.RemoveAll(dir)
		defer db.Close()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir), HasAuth: true}
		Convey("Given a HTTP request for /api/x/refreshToken while containing a valid token", func() {
			router := httprouter.New()
			router.POST("/api/x/refreshToken", middleware.AuthRequired(routing.RefreshToken(env), env))
			req := httptest.NewRequest("POST", "/api/x/refreshToken", nil)
			expireTime := time.Now().Add(time.Minute * time.Duration(utils.StandardTokenExpiration)).Unix()
			token, _ := utils.GenerateJWTToken("uuid", expireTime)
			req.Header.Add("Authorization", "bearer "+token)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("The response should return a new token with a extended expiration and add the old one to blacklist", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				tokenString := gjson.ParseBytes(body).Get("token").String()
				claims, _ := utils.ValidateJWTToken(tokenString)
				So(db.One("Token", token, &models.BlackListedToken{}), ShouldResemble, nil)
				So(claims.ExpiresAt, ShouldBeGreaterThan, expireTime)
			})
		})

		Convey("Given a HTTP request for /api/x/refreshToken while containing a invalid token", func() {
			router := httprouter.New()
			router.POST("/api/x/refreshToken", middleware.AuthRequired(routing.RefreshToken(env), env))
			req := httptest.NewRequest("POST", "/api/x/refreshToken", nil)
			token := "garbage"
			req.Header.Add("Authorization", "bearer "+token)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Should return unauthorized", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				So(string(body), ShouldResemble, "Unauthorized\n")
			})
		})

		Convey("Given a HTTP request for /api/x/refreshToken while authEnabled is false", func() {
			router := httprouter.New()
			env.HasAuth = false
			router.POST("/api/x/refreshToken", middleware.AuthRequired(routing.RefreshToken(env), env))
			req := httptest.NewRequest("POST", "/api/x/refreshToken", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Convey("Should return unauthorized", func() {
				result := resp.Result()
				So(result.StatusCode, ShouldResemble, 200)
			})
		})
	})
}
