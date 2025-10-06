package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func init() {
	root.AddCommand(create, delete)
}

type Err struct {
	Err string
}

var create = &cobra.Command{
	Use: "create",
	Args: func(cmd *cobra.Command, args []string) error {
		form := args[0]
		if form == "" {
			return errors.New("what")
		}

		_, err := os.Stat(form)
		if err != nil {
			return err
		}

		return nil
	},
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		form := args[0]

		_, err := os.Open(form)
		if err != nil {
			return err
		}

		data, err := os.ReadFile(form)
		if err != nil {
			return err
		}

		if !json.Valid(data) {
			return errors.New("invalid JSON")
		}

		url, email, password, err := basic()
		if err != nil {
			return err
		}

		res, err := client.R().SetBasicAuth(email, password).
			SetBody(data).
			SetHeader("Content-Type", "application/json").
			Post(url + "/admin/form")
		if err != nil {
			return err
		}

		if res.Err != nil {
			return res.Err
		}

		if res.StatusCode() != http.StatusOK {
			rawr := new(Err)
			val := []byte{}

			if _, err := res.Body.Read(val); err != nil {
				return err
			}

			if err := json.Unmarshal(val, rawr); err != nil {
				return err
			}

			return errors.New(rawr.Err)

		}

		log.Info("Succes")
		return nil
	},
}

var delete = &cobra.Command{
	Use: "rm",
	Args: func(cmd *cobra.Command, args []string) error {
		if args[0] == "" {
			return errors.New("")
		}

		return nil
	},
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		uri, username, password, err := basic()
		if err != nil {
			return err
		}

		url := fmt.Sprintf(uri+"/admin/form/%s", name)
		res, err := client.R().SetBasicAuth(username, password).Delete(url)
		if err != nil {
			return err
		}

		if res.Err != nil {
			return res.Err
		}

		if res.StatusCode() != http.StatusOK {
			return errors.New("")
		}

		log.Info("Succes")
		return nil
	},
}
