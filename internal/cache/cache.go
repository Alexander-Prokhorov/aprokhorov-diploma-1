package cache

import "time"

type AuthData struct {
	Login      string
	Token      string
	LastActive time.Time
}

type AuthCache interface {
	StoreToken(login string, token string) error
	GetTokenUser(token string) (string, error)
	HouseKeeper() error
	GetLifetime() time.Duration
}
