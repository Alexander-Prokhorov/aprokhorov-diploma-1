package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"aprokhorov-diploma-1/internal/cache"
	"aprokhorov-diploma-1/internal/hasher"
	"aprokhorov-diploma-1/internal/logger"
	"aprokhorov-diploma-1/internal/storage"
)

func Register(s storage.Storage, ac cache.AuthCache, hasher hasher.Hasher, log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const parent = "handlers:Register"

		// Check Headers
		log.Debug(parent, "New Request")
		if r.Header.Get("Content-Type") != "application/json" {
			log.Info(parent, "Request not 'application\\json'")
			errorText := fmt.Sprintf("only application/json supported, get %s", r.Header.Get("Content-Type"))
			http.Error(w, errorText, http.StatusNotImplemented)
			return
		}
		w.Header().Set("Content-Type", "application/json")

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

		// Authorize User in AuthCache
		log.Debug(parent, fmt.Sprintf("Try to authorize User: %s", jsonUser.Login))
		err = authorize(ac, jsonUser.Login, jsonUser.PassHash)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Check Auth after Authorizing
		auth, err := checkAuth(ac, jsonUser.Login, jsonUser.PassHash)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !auth {
			log.Error(parent, "Unexceptable behavior, user not authorized")
			http.Error(w, "Unexceptable behavior, user not authorized", http.StatusInternalServerError)
			return
		}
		log.Debug(parent, fmt.Sprintf("Successfully authorize User: %s", jsonUser.Login))

		// Respond to User
		_, err = w.Write([]byte(`{"result":"success"}`))
		if err != nil {
			log.Error(parent, "Cannot Write Respond")
		}
	}
}

func Authorize(w http.ResponseWriter, r *http.Request) {

}

func authorize(ac cache.AuthCache, login string, passHash string) error {
	err := ac.StoreSign(login, passHash)
	return err
}

func checkAuth(ac cache.AuthCache, login string, passHash string) (bool, error) {
	result, err := ac.VerifySign(login, passHash)
	return result, err
}
