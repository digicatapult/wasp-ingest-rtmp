package util

import (
	"os"

	"github.com/pkg/errors"
)

// CheckAndCreate will check if a folder path exists, otherwise create it (including parents)
func CheckAndCreate(path string) error {
	const dirPerms = 0o750

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, dirPerms)
		if err != nil {
			return errors.Wrap(err, "folder not created")
		}
	}

	return nil
}
