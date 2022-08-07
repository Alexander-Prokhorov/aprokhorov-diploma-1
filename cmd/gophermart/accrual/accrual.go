package accrual

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type Order struct {
	OrderID string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func (o Order) String() string {
	return fmt.Sprintf("OrderID:%s, Status:%s, Accrual: %v", o.OrderID, o.Status, o.Accrual)
}

type AccrualService struct {
	URL      string
	Request  *resty.Request
	Interval time.Duration
}

// `NewAccrualService` is a function that takes a URL and an interval and returns a pointer to an
// AccrualService struct
func NewAccrualService(url string, ival time.Duration) *AccrualService {
	request := resty.New().SetHeader("Context-Type", "application/json").R()
	return &AccrualService{URL: url, Request: request, Interval: ival}
}

// A function that is used to fetch data from the server.
func (a *AccrualService) FetchData(orderNo string) (Order, error) {
	order := Order{}

	// build url for request
	url := fmt.Sprintf("/api/orders/%s", orderNo)

	//fmt.Println("http://" + a.URL + url)
	// Request himself
	respond, err := a.Request.Get(a.URL + url)
	if err != nil {
		return Order{}, err
	}

	if respond.StatusCode() != http.StatusOK {
		return Order{}, errors.New("respond status not success")
	}

	err = json.Unmarshal(respond.Body(), &order)
	if err != nil {
		return Order{}, err
	}

	return order, nil
}

func (a *AccrualService) RenewOrderData(orderNo string) error {
	return nil
}
