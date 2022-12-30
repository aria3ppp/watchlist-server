package testutils

import (
	"fmt"
	"time"
)

func WaitUntil(
	f func() (done bool, err error),
	timeout, cooldown time.Duration,
) error {
	t := time.After(timeout)
	for {
		select {
		case <-t:
			return fmt.Errorf(
				"testutils.WaitUntil: %s timeout exceeded",
				timeout,
			)
		default:
			if ok, err := f(); err != nil {
				return err
			} else if ok {
				return nil
			}
		}
		time.Sleep(cooldown)
	}
}
