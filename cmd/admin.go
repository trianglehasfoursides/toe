package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.etcd.io/bbolt"
)

type detail struct {
	URL      string `json:"url"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func basic() (url string, email string, password string, err error) {
	dtl := new(detail)

	if err := db.View(func(tx *bbolt.Tx) error {
		val := tx.Bucket([]byte("default")).Get([]byte("toe"))

		fmt.Println("val", string(val))
		if val == nil {
			return errors.New("")
		}

		if err := json.Unmarshal(val, dtl); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return "", "", "", err
	}

	return dtl.URL, dtl.Email, dtl.Password, err
}
