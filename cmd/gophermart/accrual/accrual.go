package accrual

import (
	"aprokhorov-diploma-1/internal/logger"
	"encoding/json"
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
	log      logger.Logger
}

// `NewAccrualService` is a function that takes a URL and an interval and returns a pointer to an
// AccrualService struct
func NewAccrualService(url string, ival time.Duration) *AccrualService {
	request := resty.New().SetHeader("Context-Type", "application/json").R()
	log, _ := logger.NewZeroLogger("debug")
	return &AccrualService{
		URL:      url,
		Request:  request,
		Interval: ival,
		log:      log,
	}
}

// A function that is used to fetch data from the server.
func (a *AccrualService) FetchData(orderNo string) (Order, error) {
	order := Order{}

	// build url for request
	url := fmt.Sprintf("/api/orders/%s", orderNo)
	a.log.Debug("AccrualSerice", "http://"+a.URL+url)
	// Request himself
	respond, err := a.Request.Get(a.URL + url)
	if err != nil {
		return Order{}, err
	}

	if respond.StatusCode() != http.StatusOK {
		return Order{}, fmt.Errorf("respond status not success, status:%d body:%s", respond.StatusCode(), string(respond.Body()))
	}

	err = json.Unmarshal(respond.Body(), &order)
	if err != nil {
		return Order{}, err
	}

	return order, nil
}
