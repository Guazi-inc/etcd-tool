package utils

import (
	"time"

	"github.com/sirupsen/logrus"
)

func Retries(times int, sleeps time.Duration, callback func(int) error) error {
	return retriesInternal(times, 1, sleeps, callback)
}

func retriesInternal(totalTimes int, triedTimes int, sleeps time.Duration, callback func(int) error) error {
	err := callback(triedTimes)
	if err == nil {
		return nil
	} else {
		if triedTimes >= totalTimes {
			logrus.Errorf("retried %d times, all failed, last err = %s", triedTimes, err.Error())
			return err
		} else {
			logrus.Errorf("retried %d times, got err = %s",
				triedTimes, err.Error())
			time.Sleep(sleeps)
			return retriesInternal(totalTimes, triedTimes+1, sleeps, callback)
		}
	}
}
