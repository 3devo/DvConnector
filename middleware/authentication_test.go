package middleware_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/asdine/storm"
	validator "gopkg.in/go-playground/validator.v9"

	"github.com/3devo/feconnector/middleware"
	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/utils"
	"github.com/julienschmidt/httprouter"

	. "github.com/smartystreets/goconvey/convey"
)

func routeHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}

func PrepareDb() (string, *storm.DB) {
	dir := filepath.Join(os.TempDir(), "feconnector-test")
	os.MkdirAll(dir, os.ModePerm)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	return dir, db
}

func TestAuthMiddleware(t *testing.T) {
	Convey("Setup", t, func() {
		dir, db := PrepareDb()
		router := httprouter.New()
		env := &utils.Env{Db: db, Validator: validator.New(), DataDir: path.Dir(dir)}
		req := httptest.NewRequest("GET", "/test", nil)
		resp := httptest.NewRecorder()

		router.GET("/test", middleware.AuthRequired(routeHandler(), env))

		defer db.Close()
		defer os.RemoveAll(dir)

		Convey("Request /test without authorization access", func() {
			router.ServeHTTP(resp, req)
			Convey("Should respond with unauthorized", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				So(result.StatusCode, ShouldEqual, http.StatusUnauthorized)
				So(string(body), ShouldResemble, http.StatusText(http.StatusUnauthorized)+"\n")
			})
		})

		Convey("Request /test with a valid authorization header", func() {
			token, _ := utils.GenerateJWTToken("uuid", time.Now().Add(time.Hour*time.Duration(1)).Unix())
			req.Header.Set("Authorization", "bearer "+token)
			router.ServeHTTP(resp, req)
			Convey("Should respond with OK and let the request go through", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				So(result.StatusCode, ShouldEqual, http.StatusOK)
				So(string(body), ShouldEqual, "ok")
			})
		})

		Convey("request /test with a black listed token in header", func() {
			expiration := time.Now().Add(time.Hour * time.Duration(1)).Unix()
			token, _ := utils.GenerateJWTToken("uuid", expiration)
			req.Header.Set("Authorization", "bearer "+token)
			db.Save(&models.BlackListedToken{
				Token:      token,
				Expiration: expiration})
			router.ServeHTTP(resp, req)
			Convey("Should respond with unauthorized ", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				So(result.StatusCode, ShouldEqual, http.StatusUnauthorized)
				So(string(body), ShouldResemble, http.StatusText(http.StatusUnauthorized)+"\n")
			})
		})

		Convey("request /test with an expired token in header", func() {
			expiration := time.Now().Add(time.Hour * time.Duration(-1)).Unix()
			token, _ := utils.GenerateJWTToken("uuid", expiration)
			req.Header.Set("Authorization", "bearer "+token)
			router.ServeHTTP(resp, req)
			Convey("Should respond with unauthorized ", func() {
				result := resp.Result()
				body, _ := ioutil.ReadAll(result.Body)
				So(result.StatusCode, ShouldEqual, http.StatusUnauthorized)
				So(string(body), ShouldResemble, http.StatusText(http.StatusUnauthorized)+"\n")
			})
		})
	})
}
