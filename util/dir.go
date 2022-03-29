package util

import (
	"os"
)

// CheckAndCreate
func CheckAndCreate(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0o777)
		if err != nil {
			return err
		}
	}

	return nil
}
