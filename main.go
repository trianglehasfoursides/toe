package main

import (
	"fmt"
	"net"

	"github.com/charmbracelet/keygen"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	valid = validator.New()

	var err error
	if db, err = gorm.Open(sqlite.Open("toe.db"), &gorm.Config{}); err != nil {
		log.Fatal(err.Error())
	}

	if err := db.AutoMigrate(&Admin{}, &Form{}, &Question{}, &Option{}, &Answer{}); err != nil {
		log.Fatal(err.Error())
	}

	server := new(errgroup.Group)

	admn, err := admin()
	if err != nil {
		log.Fatal(err.Error())
	}

	// ssh
	server.Go(func() error {
		kp, err := keygen.New("awesome", keygen.WithPassphrase("halo"), keygen.WithKeyType(keygen.Ed25519))
		if err != nil {
			return err
		}

		server, err := wish.NewServer(
			wish.WithAddress(net.JoinHostPort("localhost", "7000")),
			wish.WithHostKeyPEM(kp.RawPrivateKey()),
			wish.WithMiddleware(
				bubbletea.Middleware(WishForm),
				func(next ssh.Handler) ssh.Handler {
					return func(s ssh.Session) {
						title := s.Command()[0]
						form := new(Form)
						if err := db.
							Preload("Questions").
							Preload("Questions.Answers").
							Preload("Questions.Options").
							Where("title = ?", title).
							First(form).Error; err != nil {
							fmt.Fprintln(s, err.Error())
							return
						}

						fmt.Fprintln(s, form.Questions[0].Text)

						s.Context().SetValue("form", form)
						next(s)
					}
				},
				func(next ssh.Handler) ssh.Handler {
					return func(s ssh.Session) {
						if arg := s.Command()[0]; arg == "" {
							fmt.Fprintln(s, "form title can't be empty")
							return
						}
						next(s)
					}
				},
				activeterm.Middleware(),
				logging.Middleware(),
			),
		)

		if err != nil {
			return err
		}

		return server.ListenAndServe()
	})

	// http
	server.Go(func() error {
		route := gin.Default()

		authorized := route.Group("/admin", gin.BasicAuth(gin.Accounts{
			admn.Email: admn.Password,
		}))

		authorized.POST("/form", CreateForm)
		authorized.DELETE("/form/:name", DeleteForm)

		return route.Run()
	})

	if err := server.Wait(); err != nil {
		log.Fatal(err.Error())
	}
}
