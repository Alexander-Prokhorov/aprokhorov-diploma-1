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

		log.Debug(parent, fmt.Sprintf("Current Balance: %f, Withdrawals: %f", balance.CurrentScore, balance.TotalWithdrawals))

		json, err := json.Marshal(balance)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(json)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Info(parent, "Send respond Successfully")
	}
}

func GetWithdrawals(s storage.Storage, log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const parent string = "handlers:GetWithdrawals"

		var userLogin loginType = "login"
		login := r.Context().Value(userLogin)
		l, ok := login.(string)
		if !ok {
			log.Error(parent, fmt.Sprintf("%v", login))
			log.Error(parent, "Login is not string")
			http.Error(w, "Login is not string", http.StatusInternalServerError)
			return
		}
		withdrawals, err := s.GetWithdrawals(r.Context(), l)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json, err := json.Marshal(withdrawals)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(json)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Info(parent, "Send respond Successfully")
	}
}

func AddWithdraw(s storage.Storage, log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const parent string = "handlers:AddWithdraw"

		var userLogin loginType = "login"
		login := r.Context().Value(userLogin)
		l, ok := login.(string)
		if !ok {
			log.Error(parent, fmt.Sprintf("%v", login))
			log.Error(parent, "Login is not string")
			http.Error(w, "Login is not string", http.StatusInternalServerError)
			return
		}

		var jsonWithdraw storage.Withdraw
		log.Debug(parent, "Parse Json")
		if err := json.NewDecoder(r.Body).Decode(&jsonWithdraw); err != nil {
			log.Info(parent, err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Info(parent, fmt.Sprintf("%v", jsonWithdraw))

		log.Debug(parent, fmt.Sprintf("Add withdraw: %f, order: %s, user: %s", jsonWithdraw.Withdraw, jsonWithdraw.OrderID, l))
		err := s.AddWithdraw(r.Context(), l, jsonWithdraw.OrderID, jsonWithdraw.Withdraw)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Info(parent, "Add withdraw Successfully")

		balance, err := s.GetBalance(r.Context(), l)
		if err != nil {
			log.Info(parent, err.Error())
		}

		log.Debug(parent, fmt.Sprintf("Current Balance: %f, Withdrawals: %f", balance.CurrentScore, balance.TotalWithdrawals))
		newBalance := balance.CurrentScore - jsonWithdraw.Withdraw
		newWithdraw := balance.TotalWithdrawals + jsonWithdraw.Withdraw

		// Check for Balance
		if newBalance < 0 {
			log.Info(parent, fmt.Sprintf("Not enought score to withdraw, Current Balance: %f, Expected Withdraw: %f", balance.CurrentScore, jsonWithdraw.Withdraw))
			http.Error(w, err.Error(), http.StatusPaymentRequired)
			return
		}

		// Update Balance
		err = s.ModifyBalance(r.Context(), l, newBalance, newWithdraw)
		if err != nil {
			log.Info(parent, err.Error())
		}
		log.Info(parent, "Recalc Balance Successfully")
		log.Debug(parent, fmt.Sprintf("New Balance: %f, Withdrawals: %f", newBalance, newWithdraw))

		respond := []byte(`{"status": "success"}`)
		_, err = w.Write(respond)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
