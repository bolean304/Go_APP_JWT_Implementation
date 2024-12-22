package utils

import (
	"errors"
	"golang-jwtauth/models"
	"regexp"
	"unicode"
)

func ValidateSignUpData(user *models.User) error {
	if user.First_name == "" || user.Last_name == "" {
		return errors.New("First Name & Last Name  Required")
	}
	if !IsValidEmail(user.Email) {
		return errors.New("Invalid Email")
	}
	if !IsStrongPassword(user.Password) {
		return errors.New("Weak Password, Make a strong Password")
	}
	return nil
}

// isValidEmail validates an email using a regex
func IsValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

// isStrongPassword validates if a password is strong
func IsStrongPassword(password string) bool {
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	var minLength = 8

	if len(password) < minLength {
		return false
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}
