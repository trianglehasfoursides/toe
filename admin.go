package main

import (
	"errors"

	"github.com/charmbracelet/huh"
	"gorm.io/gorm"
)

type Admin struct {
	ID       uint
	Email    string
	Password string
}

func admin() (*Admin, error) {
	a := &Admin{}

	if err := db.First(a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Email").
						Value(&a.Email).
						Validate(func(s string) error {
							if s == "" {
								return errors.New("Email tidak boleh kosong")
							}
							return nil
						}),
					huh.NewInput().
						Title("Password").
						Value(&a.Password).
						Validate(func(s string) error {
							if s == "" {
								return errors.New("Password tidak boleh kosong")
							}
							return nil
						}),
				),
			)

			if err := form.Run(); err != nil {
				return nil, err
			}

			if err := db.Create(a).Error; err != nil {
				return nil, err
			}

			return a, nil
		}
		return nil, err
	}

	return a, nil
}
