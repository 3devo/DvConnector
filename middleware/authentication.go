package middleware

import (
	"context"
	"net/http"
	"regexp"

	"github.com/julienschmidt/httprouter"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/utils"
)

// AuthRequired is a middleware handler that makes sure the request
// Contains a bearer authorization token and places the userId in the context if the token is valid
func AuthRequired(h httprouter.Handle, env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if env.HasAuth {
			regex := regexp.MustCompile("bearer (.*)")
			token := regex.FindStringSubmatch(r.Header.Get("Authorization"))
			if len(token) > 0 {
				blacklist := models.BlackListedToken{}
				if err := env.Db.One("Token", token[1], &blacklist); err == nil {
					http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				}
				claims, err := utils.ValidateJWTToken(token[1])
				if err != nil {
					http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), "userId", claims.Id)
				ctx = context.WithValue(r.Context(), "expiration", claims.ExpiresAt)
				r = r.WithContext(ctx)
			} else {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}
		h(w, r, ps)
	}
}
