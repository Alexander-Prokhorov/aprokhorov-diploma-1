package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"aprokhorov-diploma-1/internal/cache"
	"aprokhorov-diploma-1/internal/logger"
)

type loginType string

func CheckHeaders(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const parent = "middleware:checkContentType"
			if r.Method == http.MethodPost {
				if r.Header.Get("Content-Type") != "application/json" {
					log.Info(parent, "Request not 'application/json'")
					errorText := fmt.Sprintf("only application/json supported, get %s", r.Header.Get("Content-Type"))
					http.Error(w, errorText, http.StatusNotImplemented)
					return
				}
			}
			w.Header().Set("Content-Type", "application/json")
			log.Info(parent, "Check 'application/json' successfully, and set to")
			next.ServeHTTP(w, r)
		})
	}
}

func AuthMiddleware(ac cache.AuthCache, log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const parent string = "Middleware:Auth"

			reqToken, err := r.Cookie("GOPHER_MARKET_AUTH")
			log.Debug(parent, fmt.Sprintf("%v", r.Cookies()))
			if err != nil {
				log.Info(parent, err.Error())
				log.Info(parent, "Unauthorized request, token missed")
				http.Error(w, `{"result":"Please, Log In"}`, http.StatusUnauthorized)
				return
			}

			login, err := ac.GetTokenUser(reqToken.Value)
			if err != nil {
				log.Info(parent, err.Error())
				http.Error(w, `{"result":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}
			if login == "" {
				log.Info(parent, fmt.Sprintf("Can't find Login for token: %s", reqToken.Value))
				http.Error(w, `{"result":"Unauthorized request, token invalid or expired"}`, http.StatusUnauthorized)
				return
			}

			//Store Login in Context for user in Handlers
			var userLogin loginType = "login"
			r = r.WithContext(context.WithValue(r.Context(), userLogin, login))

			// Update Cookie to Refresh "Expires"
			//init the loc
			loc, _ := time.LoadLocation("Europe/Moscow")
			//set timezone
			expires := time.Now().In(loc).Add(ac.GetLifetime())
			newCookie := http.Cookie{Name: "GOPHER_MARKET_AUTH", Value: reqToken.Value, Expires: expires}
			http.SetCookie(w, &newCookie)

			next.ServeHTTP(w, r)

			// refresh timeout in AuthCache
			err = ac.StoreToken(login, reqToken.Value)
			if err != nil {
				log.Error(parent, err.Error())
			}
		})
	}
}
