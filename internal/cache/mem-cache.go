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
	DB      map[string]AuthData
	log     logger.Logger
	mutex   *sync.RWMutex
	timeout time.Duration
}

func NewMemCache(timeout time.Duration, log logger.Logger) *MemCache {
	db := make(map[string]AuthData)
	return &MemCache{
		DB:      db,
		log:     log,
		mutex:   &sync.RWMutex{},
		timeout: timeout,
	}
}

func (mc *MemCache) StoreSign(login string, sign string) error {
	ad := AuthData{
		Login:      login,
		Sign:       sign,
		LastActive: time.Now(),
	}

	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.DB[ad.Login] = ad
	mc.log.Debug(parent, fmt.Sprintf("Store Sign for User: %s", ad.Login))
	return nil
}

func (mc *MemCache) VerifySign(login string, sign string) (bool, error) {
	ad := AuthData{
		Login:      login,
		Sign:       sign,
		LastActive: time.Now(),
	}

	adLocal, exist := mc.DB[ad.Login]
	if !exist {
		mc.log.Debug(parent, fmt.Sprintf("Verify() No Sign for User: %s", ad.Login))
		return false, errors.New("Please, Log in")
	}

	if time.Since(ad.LastActive) > mc.timeout {
		mc.log.Debug(parent, fmt.Sprintf("Verify() Sign Expired for User: %s", ad.Login))
		return false, errors.New("Expired. Please, Login")
	}
	if ad.Sign == adLocal.Sign {
		mc.log.Debug(parent, fmt.Sprintf("Verify() Successfully find Sign for User: %s", ad.Login))
		return true, nil
	}
	mc.log.Debug(parent, fmt.Sprintf("Verify() Unexpected condition User: %s", ad.Login))
	return false, errors.New("Failed")
}

func (mc *MemCache) HouseKeeper() error {
	mc.log.Debug(parent, "HouseKeeper() Starts")
	for _, ad := range mc.DB {
		if time.Since(ad.LastActive) > mc.timeout {
			mc.mutex.Lock()
			delete(mc.DB, ad.Login)
			mc.mutex.Unlock()
			mc.log.Debug(parent, fmt.Sprintf("HouseKeeper() Delete Expired User: %s", ad.Login))
		}
	}
	return nil
}
