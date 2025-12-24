package service

import "time"

// simple retry
func Retry(times int, fn func() error) error {
	var err error
	for i := 0; i < times; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return err
}
