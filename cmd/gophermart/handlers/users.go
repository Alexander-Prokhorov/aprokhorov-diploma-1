package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aprokhorov-diploma-1/internal/cache"
	"aprokhorov-diploma-1/internal/hasher"
	"aprokhorov-diploma-1/internal/logger"
	"aprokhorov-diploma-1/internal/storage"
)

func Authorize(register bool, s storage.Storage, ac cache.AuthCache, hasher hasher.Hasher, log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const parent = "handlers:Authorize"

		log.Debug(parent, "New Request") // Migrate to middleware Access.log

		// Parse JSON from Request
		var jsonUser storage.User
		log.Debug(parent, "Parse Json")
		if err := json.NewDecoder(r.Body).Decode(&jsonUser); err != nil {
			log.Info(parent, err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Generate hash with Random Key for password
		key, err := hasher.RandomKey()
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonUser.PassHash = hasher.GetHash(jsonUser.Password, key)

		if register {
			// Register User in Database
			log.Debug(parent, fmt.Sprintf("Try to Register User: %s", jsonUser.Login))
			if err := s.RegisterUser(r.Context(), jsonUser.Login, jsonUser.PassHash, key); err != nil {
				if strings.Contains(err.Error(), "(SQLSTATE 23505)") {
					log.Info(parent, fmt.Sprintf("Already exists User with Login: %s", jsonUser.Login))
					http.Error(w, `{"result":"Login Already Used, choose another"}`, http.StatusConflict)
					return
				}
				log.Error(parent, err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Debug(parent, fmt.Sprintf("Successfully Register User: %s", jsonUser.Login))

			// Create Start Balance for User
			log.Debug(parent, fmt.Sprintf("Try to create Balance for User: %s", jsonUser.Login))
			if err := s.AddBalance(r.Context(), jsonUser.Login, 0, 0); err != nil {
				log.Error(parent, err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Debug(parent, fmt.Sprintf("Successfully created Balance for User: %s", jsonUser.Login))
		} else {
			// If not register, then validate login/pass pair from Storage
			pass, err := validatePass(r.Context(), s, hasher, jsonUser.Login, jsonUser.Password)
			if err != nil {
				log.Error(parent, err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !pass {
				log.Info(parent, fmt.Sprintf("Failed Login Attempt on Login: %s", jsonUser.Login))
				http.Error(w, `{"result":"Bad login/password"}`, http.StatusUnauthorized)
				return
			}
		}

		// Authorize User in AuthCache
		log.Debug(parent, fmt.Sprintf("Try to authorize User: %s", jsonUser.Login))

		token, err := hasher.RandomKey()
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = ac.StoreToken(jsonUser.Login, token)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Check Auth after Authorizing
		login, err := ac.GetTokenUser(token)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if login == "" {
			log.Error(parent, "Unexceptable behavior, user not authorized")
			http.Error(w, "Unexceptable behavior, user not authorized", http.StatusInternalServerError)
			return
		}
		log.Debug(parent, fmt.Sprintf("Successfully authorize User: %s", jsonUser.Login))

		// Respond to User
		// Set Auth Cookie
		//init the loc
		loc, _ := time.LoadLocation("Europe/Moscow")
		log.Info(parent, loc.String())
		//set timezone
		expires := time.Now().In(loc).Add(time.Second * 3000000)
		newCookie := http.Cookie{Name: "GOPHER_MARKET_AUTH", Value: token, Path: "/api", Expires: expires}
		http.SetCookie(w, &newCookie)

		_, err = w.Write([]byte(`{"result":"success"}`))
		if err != nil {
			log.Error(parent, "Cannot Write Respond")
		}
	}
}

func getLogin(r *http.Request, ac cache.AuthCache) string {
	requestCookies, _ := r.Cookie("GOPHER_MARKET_AUTH")
	login, _ := ac.GetTokenUser(requestCookies.Value)
	return login
}

func validatePass(ctx context.Context, s storage.Storage, hash hasher.Hasher, login string, password string) (bool, error) {
	user, err := s.GetUser(ctx, login)
	if err != nil {
		return false, err
	}

	passHash := hash.GetHash(password, user.Key)

	if passHash != user.PassHash {
		return false, nil
	}

	return true, nil
}
