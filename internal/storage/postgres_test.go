package storage

import (
	"context"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPostgres_InitTables(t *testing.T) {
	tests := []struct {
		name        string
		sqlExpected []string
	}{
		{
			name: "Create Tables",
			sqlExpected: []string{
				"CREATE TABLE IF NOT EXISTS Users \\( login text PRIMARY KEY, pass_hash text NOT NULL, key text NOT NULL, last_login timestamp NOT NULL \\)",
				"CREATE TABLE IF NOT EXISTS Balance \\( login text PRIMARY KEY, cur_score double precision NOT NULL, total_wd double precision NOT NULL \\)",
				"CREATE TABLE IF NOT EXISTS Orders \\( order_id bigint PRIMARY KEY, login text NOT NULL, status text NOT NULL, score double precision NOT NULL, created_at timestamp NOT NULL, last_changed timestamp NOT NULL \\)",
				"CREATE TABLE IF NOT EXISTS Withdrawals \\( order_id bigint PRIMARY KEY, login text NOT NULL, wd double precision NOT NULL, time timestamp NOT NULL \\)",
			},
		},
	}

	for _, tt := range tests {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		t.Run(tt.name, func(t *testing.T) {
			p := &Postgres{
				DB:         db,
				mutex:      &sync.RWMutex{},
				Statements: Statements{},
			}

			for _, query := range tt.sqlExpected {
				mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))
			}
			err := p.InitTables(context.Background())
			assert.NoError(t, err)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestPostgres_PrepareStatemets(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "Create Statements Test",
			want: []string{
				`INSERT INTO Users \(login, pass_hash, key, last_login\) VALUES \(\$1, \$2, \$3, \$4\)`,
				`SELECT login, pass_hash, key, last_login FROM Users WHERE login = \$1`,
				`SELECT login, pass_hash, key, last_login FROM Users`,
				`INSERT INTO Orders \(order_id, login, status, score, created_at, last_changed\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`,
				`UPDATE Orders SET status = \$2, score = \$3, last_changed = \$4 WHERE order_id = \$1`,
				`SELECT order_id, login, status, score, last_changed, created_at FROM Orders WHERE order_id = \$1`,
				`SELECT order_id, login, status, score, last_changed, created_at FROM Orders WHERE login = \$1`,
				`SELECT order_id, login, status, score, last_changed, created_at FROM Orders WHERE status != 'INVALID' AND status != 'PROCESSED'`,
				`INSERT INTO Balance \(login, cur_score, total_wd\) VALUES \(\$1, \$2, \$3\)`,
				`UPDATE Balance SET cur_score = \$2, total_wd = \$3 WHERE login = \$1`,
				`SELECT login, cur_score, total_wd FROM Balance WHERE login = \$1`,
				`INSERT INTO Withdrawals \(order_id, login, wd, time\) VALUES \(\$1, \$2, \$3, \$4\)`,
				`SELECT order_id, login, wd, time FROM Withdrawals WHERE login = \$1`,
			},
		},
	}
	for _, tt := range tests {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		t.Run(tt.name, func(t *testing.T) {
			p := &Postgres{
				DB:         db,
				mutex:      &sync.RWMutex{},
				Statements: Statements{},
			}
			for _, expStmt := range tt.want {
				mock.ExpectPrepare(expStmt)
			}

			err := p.PrepareStatements(context.Background())
			assert.NoError(t, err)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestPostgres_RegisterUser(t *testing.T) {

	type args struct {
		login string
		hash  string
		key   string
	}
	type want struct {
		rowInsertID int
		rowAffected int
	}
	tests := []struct {
		name string
		args []args
		want want
	}{
		{
			name: "Test Insert User1",
			args: []args{
				{login: "User1", hash: "hash1", key: "key"},
			},
			want: want{
				rowInsertID: 1,
				rowAffected: 1,
			},
		},
	}

	for _, tt := range tests {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		t.Run(tt.name, func(t *testing.T) {
			p := &Postgres{
				DB:         db,
				mutex:      &sync.RWMutex{},
				Statements: Statements{},
			}

			ctx := context.Background()

			preparation := []string{
				`INSERT INTO Users \(login, pass_hash, key, last_login\) VALUES \(\$1, \$2, \$3, \$4\)`,
				`SELECT login, pass_hash, key, last_login FROM Users WHERE login = \$1`,
				`SELECT login, pass_hash, key, last_login FROM Users`,
				`INSERT INTO Orders \(order_id, login, status, score, created_at, last_changed\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`,
				`UPDATE Orders SET status = \$2, score = \$3, last_changed = \$4 WHERE order_id = \$1`,
				`SELECT order_id, login, status, score, last_changed, created_at FROM Orders WHERE order_id = \$1`,
				`SELECT order_id, login, status, score, last_changed, created_at FROM Orders WHERE login = \$1`,
				`SELECT order_id, login, status, score, last_changed, created_at FROM Orders WHERE status != 'INVALID' AND status != 'PROCESSED'`,
				`INSERT INTO Balance \(login, cur_score, total_wd\) VALUES \(\$1, \$2, \$3\)`,
				`UPDATE Balance SET cur_score = \$2, total_wd = \$3 WHERE login = \$1`,
				`SELECT login, cur_score, total_wd FROM Balance WHERE login = \$1`,
				`INSERT INTO Withdrawals \(order_id, login, wd, time\) VALUES \(\$1, \$2, \$3, \$4\)`,
				`SELECT order_id, login, wd, time FROM Withdrawals WHERE login = \$1`,
			}

			for _, query := range preparation {
				mock.ExpectPrepare(query)
			}

			err := p.PrepareStatements(ctx)
			assert.NoError(t, err)

			for _, args := range tt.args {
				query := `INSERT INTO Users \(login, pass_hash, key, last_login\) VALUES \(\$1, \$2, \$3, \$4\)`
				mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))

				err := p.RegisterUser(ctx, args.login, args.hash, args.key)
				assert.NoError(t, err)

				err = mock.ExpectationsWereMet()
				assert.NoError(t, err)
			}
		})
	}
}
