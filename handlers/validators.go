package handlers

import "github.com/go-playground/validator/v10"

func init() {
	validate = validator.New()
	if err := validate.RegisterValidation("password", passwordValidation); err != nil {
		panic(err)
	}
}

func passwordValidation(fl validator.FieldLevel) bool {
	pass := fl.Field().String()

	hasLowerCase := false
	hasUpperCase := false
	hasDigit := false
	for _, c := range pass {
		if c >= 'a' && c <= 'z' {
			hasLowerCase = true
		} else if c >= 'A' && c <= 'Z' {
			hasUpperCase = true
		} else if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}

	return hasUpperCase && hasLowerCase && hasDigit && len(pass) < 8 && len(pass) > 32
}
