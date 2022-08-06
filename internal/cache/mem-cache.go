package cache

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"aprokhorov-diploma-1/internal/logger"
)

const parent string = "MemCache"

type MemCache struct {
	DB       map[string]AuthData
	log      logger.Logger
	mutex    *sync.RWMutex
	Lifetime time.Duration
}

func NewMemCache(timeout time.Duration, log logger.Logger) *MemCache {
	db := make(map[string]AuthData)
	return &MemCache{
		DB:       db,
		log:      log,
		mutex:    &sync.RWMutex{},
		Lifetime: timeout,
	}
}

func (mc *MemCache) GetLifetime() time.Duration {
	return mc.Lifetime
}

func (mc *MemCache) StoreToken(login string, token string) error {
	ad := AuthData{
		Login:      login,
		Token:      token,
		LastActive: time.Now(),
	}

	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.DB[ad.Token] = ad
	mc.log.Debug(parent, fmt.Sprintf("Store Token for User: %s", ad.Login))
	return nil
}

func (mc *MemCache) GetTokenUser(token string) (string, error) {
	ad := AuthData{
		Login:      "",
		Token:      token,
		LastActive: time.Now(),
	}

	adLocal, exist := mc.DB[ad.Token]
	if !exist {
		mc.log.Debug(parent, fmt.Sprintf("Verify() No User for Token: %s", ad.Token))
		return ad.Login, errors.New("Please, Log in")
	}

	if time.Since(adLocal.LastActive) > mc.Lifetime {
		mc.log.Debug(parent, fmt.Sprintf("Verify() Token Expired for User: %s", adLocal.Login))
		return ad.Login, errors.New("Expired. Please, Login")
	}
	if ad.Token == adLocal.Token {
		mc.log.Debug(parent, fmt.Sprintf("Verify() Successfully find Token for User: %s", adLocal.Login))
		return adLocal.Login, nil
	}
	mc.log.Debug(parent, fmt.Sprintf("Verify() Unexpected condition Token: %s", ad.Token))
	return ad.Login, errors.New("Failed")
}

func (mc *MemCache) HouseKeeper() error {
	mc.log.Debug(parent, "HouseKeeper() Starts")
	for _, ad := range mc.DB {
		if time.Since(ad.LastActive) > mc.Lifetime {
			mc.mutex.Lock()
			delete(mc.DB, ad.Token)
			mc.mutex.Unlock()
			mc.log.Debug(parent, fmt.Sprintf("HouseKeeper() Delete Expired Token: %s", ad.Token))
		}
	}
	return nil
}
