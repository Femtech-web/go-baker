package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	FieldErrors    map[string]string
	NonFieldErrors []string
}

// check form submission if valid method
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

// add error field method
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// add non field error method
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// check a form field method
func (v *Validator) CheckField(pass bool, key, message string) {
	if !pass {
		v.AddFieldError(key, message)
	}
}

// not blank
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// max chars
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// min chars
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// permitted value
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}

	return false
}

// matches
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// is same
func IsSame(first, second string) bool {
	return first == second
}
