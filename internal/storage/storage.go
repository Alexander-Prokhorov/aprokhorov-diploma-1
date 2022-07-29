package storage

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

// Data Structs
type Parser interface {
	New() Parser
	Parse(params []string) error
}

type User struct {
	Login     string    `db:"login"`
	PassHash  string    `db:"pass_hash"`
	LastLogin time.Time `db:"last_login"`
}

func (u *User) New() Parser { return &User{} }

func (u *User) Parse(values []string) error {
	if values == nil {
		*u = User{}
		return nil
	}

	for i, value := range values {
		sv, err := driver.String.ConvertValue(value)
		if err != nil {
			return fmt.Errorf("cannot scan value. %w", err)
		}

		v, ok := sv.(string)
		if !ok {
			return err
		}

		// Value Order:
		// login, pass_hash, last_login
		switch i {
		case 0:
			u.Login = v
		case 1:
			u.PassHash = v
		case 2:
			time, err := time.Parse("2006-01-02T15:04:05.000000Z", v)
			if err != nil {
				return err
			}
			u.LastLogin = time
		}
	}
	return nil
}

type Order struct {
	OrderId    string    `db:"order_id"`
	Login      string    `db:"login"`
	Status     string    `db:"status"`
	Score      int       `db:"score"`
	LastChange time.Time `db:"last_changed"`
}

func (o *Order) New() Parser { return &Order{} }

func (o *Order) Parse(values []string) error {
	if values == nil {
		*o = Order{}
		return nil
	}

	for i, value := range values {
		sv, err := driver.String.ConvertValue(value)
		if err != nil {
			return fmt.Errorf("cannot scan value. %w", err)
		}

		v, ok := sv.(string)
		if !ok {
			return err
		}
		// Value Order:
		// order_id, login, status, score, last_changed
		switch i {
		case 0:
			o.OrderId = v
		case 1:
			o.Login = v
		case 2:
			o.Status = v
		case 3:
			score, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			o.Score = score
		case 4:
			time, err := time.Parse("2006-01-02T15:04:05.000000Z", v)
			if err != nil {
				return err
			}
			o.LastChange = time
		}
	}
	return nil
}

type Withdraw struct {
	OrderId  string    `db:"order_id"`
	Login    string    `db:"login"`
	Withdraw int       `db:"wd"`
	Time     time.Time `db:"time"`
}

func (w *Withdraw) New() Parser { return &Withdraw{} }

func (w *Withdraw) Parse(values []string) error {
	*w = Withdraw{}
	if values == nil {
		return nil
	}

	for i, value := range values {
		sv, err := driver.String.ConvertValue(value)
		if err != nil {
			return fmt.Errorf("cannot scan value. %w", err)
		}

		v, ok := sv.(string)
		if !ok {
			return err
		}
		// Value Order:
		// order_id, login, status, score, last_changed
		switch i {
		case 0:
			w.OrderId = v
		case 1:
			w.Login = v
		case 2:
			vv, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			w.Withdraw = vv
		case 3:
			time, err := time.Parse("2006-01-02T15:04:05.000000Z", v)
			if err != nil {
				return err
			}

			w.Time = time
		}

	}

	return nil
}

type Balance struct {
	Login            string `db:"login"`
	CurrentScore     int    `db:"cur_score"`
	TotalWithdrawals int    `db:"total_wd"`
}

func (b *Balance) New() Parser { return &Balance{} }

func (b *Balance) Parse(values []string) error {
	*b = Balance{}
	if values == nil {
		return nil
	}

	for i, value := range values {
		sv, err := driver.String.ConvertValue(value)
		if err != nil {
			return fmt.Errorf("cannot scan value. %w", err)
		}

		v, ok := sv.(string)
		if !ok {
			return err
		}
		fmt.Println("Parsing: ", v)
		// Value Order:
		// login, cur_score, total_wd
		switch i {
		case 0:
			b.Login = v
		case 1:
			vv, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			b.CurrentScore = vv
		case 2:
			vv, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			b.TotalWithdrawals = vv
		}
	}
	return nil
}

// END

// Interface for use in Project
type Storage interface {
	RegisterUser(ctx context.Context, login string, hash string) error
	GetUsers(ctx context.Context) ([]*User, error)
	AddOrder(ctx context.Context, login string, order string) error
	ModifyOrder(ctx context.Context, login string, order string, status string, score int) error
	GetOrders(ctx context.Context, login string) ([]*Order, error)
	AddBalance(ctx context.Context, login string, score int, wd int) error
	ModifyBalance(ctx context.Context, login string, score int, wd int) error
	GetBalance(ctx context.Context, login string) (Balance, error)
	AddWithdraw(ctx context.Context, login string, order string, wd int) error
	GetWithdrawals(ctx context.Context, login string) ([]*Withdraw, error)
}
