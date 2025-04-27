package models

import (
	"golang.org/x/time/rate"
	"strconv"
	"time"
)

type Limiters struct {
	RegLimiter   *rate.Limiter
	LoginLimiter *rate.Limiter
}

func newLimiters(regRate time.Duration, regBurst int, loginRate time.Duration, loginBurst int) *Limiters {

	return &Limiters{
		RegLimiter:   rate.NewLimiter(rate.Limit(regRate), regBurst),
		LoginLimiter: rate.NewLimiter(rate.Limit(loginRate), loginBurst),
	}

}

func NewLimitersByEnv(regRate string, regBurst string, loginRate string, loginBurst string) (*Limiters, error) {

	regRateI, err := time.ParseDuration(regRate)
	if err != nil {
		return nil, err
	}

	regBurstI, err := strconv.Atoi(regBurst)
	if err != nil {
		return nil, err
	}

	loginRateI, err := time.ParseDuration(loginRate)
	if err != nil {
		return nil, err
	}

	loginBurstI, err := strconv.Atoi(loginBurst)
	if err != nil {
		return nil, err
	}

	return newLimiters(regRateI, regBurstI, loginRateI, loginBurstI), nil
}
