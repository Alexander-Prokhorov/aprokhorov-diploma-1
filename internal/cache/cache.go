package cache

import "time"

type AuthData struct {
	Login      string
	Sign       string
	LastActive time.Time
}

type AuthCache interface {
	StoreSign(login string, sign string) error
	VerifySign(login string, sign string) (bool, error)
	HouseKeeper() error
}
