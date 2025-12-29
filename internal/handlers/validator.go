package handlers

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	validate = validator.New()

	// Регистрируем кастомную валидацию slug
	_ = validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		slug := fl.Field().String()
		// Регулярка для slug: только латиница, цифры, дефисы
		matched, _ := regexp.MatchString("^[a-z0-9]+(?:-[a-z0-9]+)*$", slug)
		return matched
	})
}
