package main

import (
	"encoding/json"
	"errors"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"resty.dev/v3"

	"go.etcd.io/bbolt"
)

func main() {
	client = resty.New()
	defer client.Close()

	var err error
	if db, err = bbolt.Open("db.toe", 0600, bbolt.DefaultOptions); err != nil {
		log.Fatal(err.Error())
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("default")); err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Fatal(err.Error())
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		if val := tx.Bucket([]byte("default")).Get([]byte("toe")); val == nil {
			return errors.New("")
		}

		return nil
	}); err != nil {

		dtl := new(detail)
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("URL").Value(&dtl.URL),
				huh.NewInput().Title("Email").Value(&dtl.Email),
				huh.NewInput().Title("Password").Value(&dtl.Password),
			),
		)

		form.WithTheme(huh.ThemeCatppuccin())

		if err := form.Run(); err != nil {
			log.Fatal(err.Error())
		}

		if err := db.Update(func(tx *bbolt.Tx) error {
			val, err := json.Marshal(dtl)
			if err != nil {
				return err
			}

			return tx.Bucket([]byte("default")).Put([]byte("toe"), val)
		}); err != nil {
			log.Fatal(err.Error())
		}
	}

	fang.Execute(root.Context(), root)
}
