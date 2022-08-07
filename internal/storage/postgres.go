package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type Postgres struct {
	DB         *sql.DB
	mutex      *sync.RWMutex
	Statements Statements
}

type Statements struct {
	InsertUser         *sql.Stmt
	SelectUser         *sql.Stmt
	SelectUsers        *sql.Stmt
	InsertOrder        *sql.Stmt
	UpdateOrder        *sql.Stmt
	SelectOrder        *sql.Stmt
	SelectOrdersByUser *sql.Stmt
	SelectOrdersUndone *sql.Stmt
	InsertBalance      *sql.Stmt
	UpdateBalance      *sql.Stmt
	SelectBalance      *sql.Stmt
	InsertWithdraw     *sql.Stmt
	SelectWithdrawals  *sql.Stmt
}

func NewPostgresClient(ctx context.Context, address string, dbname string) (Postgres, error) {
	dbName := ""
	if dbname != "" {
		dbName = fmt.Sprintf("/%s", dbname)
	}
	db, err := sql.Open("pgx", address+dbName)

	// Сравнить значение err с ошибкой sql Database NOT Exist
	if err != nil {
		return Postgres{}, err
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(10)

	newPGS := Postgres{
		DB:         db,
		mutex:      &sync.RWMutex{},
		Statements: Statements{},
	}

	err = newPGS.InitTables(ctx)
	if err != nil {
		return Postgres{}, err
	}

	err = newPGS.PrepareStatements(ctx)
	if err != nil {
		return Postgres{}, err
	}

	return newPGS, nil
}

func (p *Postgres) GracefulShutdown() {
	// Close Statements

	p.Statements.InsertUser.Close()
	p.Statements.SelectUser.Close()
	p.Statements.SelectUsers.Close()
	p.Statements.InsertOrder.Close()
	p.Statements.UpdateOrder.Close()
	p.Statements.SelectOrder.Close()
	p.Statements.SelectOrdersByUser.Close()
	p.Statements.SelectOrdersUndone.Close()
	p.Statements.InsertBalance.Close()
	p.Statements.UpdateBalance.Close()
	p.Statements.SelectBalance.Close()
	p.Statements.InsertWithdraw.Close()
	p.Statements.SelectWithdrawals.Close()

	// Close DB
	p.DB.Close()
}

func (p *Postgres) InitTables(ctx context.Context) error {
	scheme := []string{
		`Users (
			login text PRIMARY KEY,
			pass_hash text NOT NULL,
			key text NOT NULL,
			last_login timestamp NOT NULL
			)`,

		`Balance (
			login text PRIMARY KEY,
			cur_score int NOT NULL,
			total_wd int NOT NULL
			)`,

		`Orders (
			order_id bigint PRIMARY KEY,
			login text NOT NULL,
			status text NOT NULL,
			score int NOT NULL,
			created_at timestamp NOT NULL,
			last_changed timestamp NOT NULL
			)`,

		`Withdrawals (
			order_id bigint PRIMARY KEY,
			login text NOT NULL,
			wd int NOT NULL,
			time timestamp NOT NULL
			)`,
	}

	for _, table := range scheme {
		query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s", table)
		_, err := p.DB.ExecContext(ctx, query)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Postgres) PrepareStatements(ctx context.Context) error {
	stmt, err := p.DB.PrepareContext(ctx, "INSERT INTO Users (login, pass_hash, key, last_login) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	p.Statements.InsertUser = stmt

	stmt, err = p.DB.PrepareContext(ctx, "SELECT login, pass_hash, key, last_login FROM Users WHERE login = $1")
	if err != nil {
		return err
	}
	p.Statements.SelectUser = stmt

	stmt, err = p.DB.PrepareContext(ctx, "SELECT login, pass_hash, key, last_login FROM Users")
	if err != nil {
		return err
	}
	p.Statements.SelectUsers = stmt

	stmt, err = p.DB.PrepareContext(ctx, "INSERT INTO Orders (order_id, login, status, score, created_at, last_changed) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return err
	}
	p.Statements.InsertOrder = stmt

	stmt, err = p.DB.PrepareContext(ctx, "UPDATE Orders SET status = $2, score = $3, last_changed = $4 WHERE order_id = $1")
	if err != nil {
		return err
	}
	p.Statements.UpdateOrder = stmt

	stmt, err = p.DB.PrepareContext(ctx, "SELECT order_id, login, status, score, last_changed, created_at FROM Orders WHERE order_id = $1")
	if err != nil {
		return err
	}
	p.Statements.SelectOrder = stmt

	stmt, err = p.DB.PrepareContext(ctx, "SELECT order_id, login, status, score, last_changed, created_at FROM Orders WHERE login = $1")
	if err != nil {
		return err
	}
	p.Statements.SelectOrdersByUser = stmt

	stmt, err = p.DB.PrepareContext(ctx, "SELECT order_id, login, status, score, last_changed, created_at FROM Orders WHERE status != 'INVALID' AND status != 'PROCESSED'")
	if err != nil {
		return err
	}
	p.Statements.SelectOrdersUndone = stmt

	stmt, err = p.DB.PrepareContext(ctx, "INSERT INTO Balance (login, cur_score, total_wd) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	p.Statements.InsertBalance = stmt

	stmt, err = p.DB.PrepareContext(ctx, "UPDATE Balance SET cur_score = $2, total_wd = $3 WHERE login = $1")
	if err != nil {
		return err
	}
	p.Statements.UpdateBalance = stmt

	stmt, err = p.DB.PrepareContext(ctx, "SELECT login, cur_score, total_wd FROM Balance WHERE login = $1")
	if err != nil {
		return err
	}
	p.Statements.SelectBalance = stmt

	stmt, err = p.DB.PrepareContext(ctx, "INSERT INTO Withdrawals (order_id, login, wd, time) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	p.Statements.InsertWithdraw = stmt

	stmt, err = p.DB.PrepareContext(ctx, "SELECT order_id, login, wd, time FROM Withdrawals WHERE login = $1")
	if err != nil {
		return err
	}
	p.Statements.SelectWithdrawals = stmt

	return nil
}

func (p Postgres) RegisterUser(ctx context.Context, login string, hash string, key string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	_, err := p.Statements.InsertUser.ExecContext(ctx, login, hash, key, time.Now())
	return err
}

func (p Postgres) GetUser(ctx context.Context, login string) (User, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	users, err := getBulk[*User](ctx, p.Statements.SelectUser, login)
	if err != nil {
		return User{}, err
	}
	if len(users) > 1 {
		return User{}, errors.New("PG: GetUser unexpected error, get more than 1 result")
	} else if len(users) == 0 {
		return User{}, sql.ErrNoRows
	}

	return *users[0], nil
}

func (p Postgres) GetUsers(ctx context.Context) ([]*User, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return getBulk[*User](ctx, p.Statements.SelectUsers)
}

func (p Postgres) AddOrder(ctx context.Context, login string, order string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	time := time.Now()
	_, err := p.Statements.InsertOrder.ExecContext(ctx, order, login, "NEW", 0, time, time)
	return err
}

func (p Postgres) ModifyOrder(ctx context.Context, order string, status string, score int) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	_, err := p.Statements.UpdateOrder.ExecContext(ctx, order, status, score, time.Now())
	return err
}

func (p Postgres) GetOrder(ctx context.Context, order string) (Order, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	orders, err := getBulk[*Order](ctx, p.Statements.SelectOrder, order)
	if err != nil {
		return Order{}, err
	}

	if len(orders) > 1 {
		return Order{}, errors.New("PG: GetOrder unexpected error, get more than 1 result")
	} else if len(orders) == 0 {
		return Order{}, sql.ErrNoRows
	}

	return *orders[0], nil
}

func (p Postgres) GetOrdersByUser(ctx context.Context, login string) ([]*Order, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return getBulk[*Order](ctx, p.Statements.SelectOrdersByUser, login)
}

func (p Postgres) GetOrdersUndone(ctx context.Context) ([]*Order, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return getBulk[*Order](ctx, p.Statements.SelectOrdersUndone)
}

func (p Postgres) AddBalance(ctx context.Context, login string, score int, wd int) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	_, err := p.Statements.InsertBalance.ExecContext(ctx, login, score, wd)
	return err
}

func (p Postgres) ModifyBalance(ctx context.Context, login string, score int, wd int) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	_, err := p.Statements.UpdateBalance.ExecContext(ctx, login, score, wd)
	return err
}

func (p Postgres) GetBalance(ctx context.Context, login string) (Balance, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	balances, err := getBulk[*Balance](ctx, p.Statements.SelectBalance, login)
	if err != nil {
		return Balance{}, err
	}
	if len(balances) > 1 {
		return Balance{}, errors.New("PG: GetBalance unexpected error, get more than 1 result")
	} else if len(balances) == 0 {
		return Balance{}, sql.ErrNoRows
	}

	return *balances[0], nil
}

func (p Postgres) AddWithdraw(ctx context.Context, login string, order string, wd int) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	_, err := p.Statements.InsertWithdraw.ExecContext(ctx, order, login, wd, time.Now())
	return err
}

func (p Postgres) GetWithdrawals(ctx context.Context, login string) ([]*Withdraw, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return getBulk[*Withdraw](ctx, p.Statements.SelectWithdrawals, login)
}

func getBulk[T Parser](ctx context.Context, stmt *sql.Stmt, args ...any) ([]T, error) {
	result := make([]T, 0)
	var rows *sql.Rows
	var err error

	rows, err = stmt.QueryContext(ctx, args...)
	if err != nil {
		return result, err
	}

	columns, err := rows.Columns()
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var target T
		target = target.New().(T)
		params := make([]string, len(columns))
		paramsPointers := make([]interface{}, len(columns))
		for i := range params {
			paramsPointers[i] = &params[i]
		}
		if err := rows.Scan(paramsPointers...); err != nil {
			return result, err
		}
		err := target.Parse(params)
		if err != nil {
			return result, err
		}
		result = append(result, target)
	}
	return result, err
}
