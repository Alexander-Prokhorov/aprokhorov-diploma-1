package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"aprokhorov-diploma-1/internal/logger"
	"aprokhorov-diploma-1/internal/storage"
)

func GetBalance(s storage.Storage, log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const parent string = "handlers:GetBalance"

		var userLogin loginType = "login"
		login := r.Context().Value(userLogin)
		l, ok := login.(string)
		if !ok {
			log.Error(parent, fmt.Sprintf("%v", login))
			log.Error(parent, "Login is not string")
			http.Error(w, "Login is not string", http.StatusInternalServerError)
			return
		}
		balance, err := s.GetBalance(r.Context(), l)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json, err := json.Marshal(balance)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(json)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Info(parent, "Send respond Successfully")
	}
}

func AddWithdraw(w http.ResponseWriter, r *http.Request) {

}

func GetWithdrawals(w http.ResponseWriter, r *http.Request) {

}
