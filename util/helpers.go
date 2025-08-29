package util

import (
	"encoding/json"
	"os"
	"time"
)

// MarshalToJSON marshals data to JSON []byte, returns error if failed
func MarshalToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// NowUTC returns current time in UTC
func NowUTC() time.Time {
	return time.Now().UTC()
}

// CreateDirectory create multiple directory.
func CreateDirectory(paths ...string) (err error) {
	for _, path := range paths {
		_, notExistError := os.Stat(path)
		if os.IsNotExist(notExistError) {
			if err = os.MkdirAll(path, os.ModePerm); err != nil {
				return err
			}
		}
	}
	return
}
