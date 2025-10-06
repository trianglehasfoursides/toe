package main

import (
	"net/http"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var validate = validator.New()

type Form struct {
	gorm.Model
	Title     string     `json:"title" validate:"required"`
	Questions []Question `json:"questions" validate:"required,dive"`
}

type Question struct {
	gorm.Model
	FormID  uint     `json:"form_id"`
	Text    string   `json:"text" validate:"required"`
	Type    string   `json:"type" validate:"required,oneof=text textarea radio checkbox"`
	Options []Option `json:"options"`
	Answers []Answer `json:"answers"`
}

type Option struct {
	gorm.Model
	QuestionID uint   `json:"question_id"`
	Text       string `json:"text" validate:"required"`
}

type Answer struct {
	gorm.Model
	FormID     uint    `json:"form_id"`
	QuestionID uint    `json:"question_id"`
	OptionID   *uint   `json:"option_id"`
	UserID     uint    `json:"user_id"`
	Text       string  `json:"text"`
	Option     *Option `json:"option,omitempty"`
}

func CreateForm(ctx *gin.Context) {
	form := new(Form)
	if err := ctx.ShouldBindJSON(form); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, q := range form.Questions {
		if len(q.Answers) > 0 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "answers tidak boleh ada saat membuat form"})
			return
		}
	}

	if err := validate.Struct(form); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Create(form).Error; err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, form)
}

func DeleteForm(ctx *gin.Context) {
	title := ctx.Param("title")
	if title == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	if err := db.Where("title = ?", title).Delete(&Form{}).Error; err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusOK)
}

func WishForm(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	form := s.Context().Value("form").(*Form)

	fields := []huh.Field{}
	answers := []Answer{}

	for _, f := range form.Questions {
		answer := Answer{
			FormID:     f.FormID,
			QuestionID: f.ID,
		}

		switch f.Type {
		case "text", "textarea":
			input := huh.NewInput().
				Title(f.Text).
				Value(&answer.Text)
			fields = append(fields, input)
			answers = append(answers, answer)

		case "checkbox":
			options := huh.NewOptions[string]()
			for _, o := range f.Options {
				options = append(options, huh.NewOption(o.Text, o.Text))
			}
			var selected []string
			input := huh.NewMultiSelect[string]().
				Title(f.Text).
				Options(options...).
				Value(&selected)
			fields = append(fields, input)

			defer func(q Question, sel *[]string) {
				for _, s := range *sel {
					for _, o := range q.Options {
						if o.Text == s {
							id := o.ID
							answers = append(answers, Answer{
								FormID:     q.FormID,
								QuestionID: q.ID,
								OptionID:   &id,
							})
						}
					}
				}
			}(f, &selected)

		case "radio":
			options := huh.NewOptions[string]()
			for _, o := range f.Options {
				options = append(options, huh.NewOption(o.Text, o.Text))
			}
			var selected string
			input := huh.NewSelect[string]().
				Title(f.Text).
				Options(options...).
				Value(&selected)
			fields = append(fields, input)

			defer func(q Question, sel *string) {
				for _, o := range q.Options {
					if o.Text == *sel {
						id := o.ID
						answers = append(answers, Answer{
							FormID:     q.FormID,
							QuestionID: q.ID,
							OptionID:   &id,
						})
					}
				}
			}(f, &selected)
		}
	}

	var confirm bool
	fields = append(fields, huh.NewConfirm().Title("Submit?").Value(&confirm))

	fr := huh.NewForm(huh.NewGroup(fields...))

	return &model{
		form:    fr,
		answers: answers,
	}, nil
}

type model struct {
	answers []Answer
	form    tea.Model // ganti jadi interface
}

func (m *model) Init() tea.Cmd {
	return m.form.Init()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newForm, cmd := m.form.Update(msg)

	// update reference
	m.form = newForm

	if f, ok := m.form.(*huh.Form); ok {
		switch f.State {
		case huh.StateCompleted:
			if len(m.answers) > 0 {
				if err := db.Create(&m.answers).Error; err != nil {
					log.Fatal(err.Error())
				}
			}
			return m, tea.Quit
		case huh.StateAborted:
			return m, tea.Quit
		}
	}

	return m, cmd
}

func (m *model) View() string {
	return m.form.View()
}
