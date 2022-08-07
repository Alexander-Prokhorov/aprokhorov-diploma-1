package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"aprokhorov-diploma-1/internal/logger"
	"aprokhorov-diploma-1/internal/storage"
	"aprokhorov-diploma-1/internal/verificator"
)

// 9278923470
// 12345678903
// 346436439
// 4100401111100062
// 371449635398431

func NewOrder(s storage.Storage, v verificator.Verificator, log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const parent string = "handlers:NewOrder"
		loginAny := r.Context().Value(loginType("login"))
		login, ok := loginAny.(string)
		if !ok {
			log.Error(parent, "Cannot get valid login from Context")
			http.Error(w, "Cannot get valid login from Context", http.StatusInternalServerError)
			return
		}
		// Check Header for context-type
		if r.Header.Get("Content-Type") != "text/plain" {
			log.Info(parent, "Request not 'text/plain'")
			errorText := fmt.Sprintf("only text/plain supported, get %s", r.Header.Get("Content-Type"))
			http.Error(w, errorText, http.StatusBadRequest)
			return
		}
		// Try to read raw order data
		orderRaw, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Info(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		r.Body.Close()
		// Convert order raw data to integer
		orderNo, err := strconv.ParseInt(string(orderRaw), 10, 64)
		if err != nil {
			log.Info(parent, err.Error())
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		// Validate order number
		valid := v.Valid(orderNo)
		if !valid {
			log.Info(parent, fmt.Sprintf("Bad order_no %v", orderNo))
			http.Error(w, fmt.Sprintf("Bad order_no %v", orderNo), http.StatusUnprocessableEntity)
			return
		}
		// Lookup in Storage for this Order and check for existance
		localOrder, err := s.GetOrder(r.Context(), string(orderRaw))
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Error(parent, err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// No Rows, Create New Order
			log.Debug(parent, fmt.Sprintf("No order %v yet, create some", orderNo))
			err := s.AddOrder(r.Context(), login, string(orderRaw))
			if err != nil {
				log.Error(parent, err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Error(w, "Success", http.StatusAccepted)
			log.Info(parent, fmt.Sprintf("Create order %v for %v successfully", orderNo, login))
			return
		}
		// If Order exist already - check why
		if !errors.Is(err, sql.ErrNoRows) {
			if localOrder.Login != login {
				log.Info(parent, fmt.Sprintf("User:%v try to upload order{%v} that have been uploaded by user:%v", login, orderNo, localOrder.Login))
				http.Error(w, fmt.Sprintf("Order: %v alredy have been uploaded by another User", orderNo), http.StatusConflict)
				return
			} else {
				log.Info(parent, fmt.Sprintf("User:%v try to upload order{%v} that have been uploaded by himself", login, orderNo))
				http.Error(w, fmt.Sprintf("Order %v already have been uploaded by you", orderNo), http.StatusOK)
				return
			}
		}
	}
}

func GetOrders(s storage.Storage, log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const parent string = "handlers:GetOrder"
		loginAny := r.Context().Value(loginType("login"))
		login, ok := loginAny.(string)
		if !ok {
			log.Error(parent, "Cannot get valid login from Context")
			http.Error(w, "Cannot get valid login from Context", http.StatusInternalServerError)
			return
		}

		orders, err := s.GetOrdersByUser(r.Context(), login)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Info(parent, fmt.Sprintf("User:%v No Orders", login))
				http.Error(w, "No Orders", http.StatusNoContent)
				return
			}
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}

		if len(orders) == 0 {
			log.Info(parent, fmt.Sprintf("User:%v No Orders", login))
			http.Error(w, "No Orders", http.StatusNoContent)
			return
		}

		ordersJSON, err := json.MarshalIndent(orders, "", "  ")
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(ordersJSON)
		if err != nil {
			log.Error(parent, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
