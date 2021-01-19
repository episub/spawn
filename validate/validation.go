package validate

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid"

	"cloud.google.com/go/civil"
)

var numberRx = regexp.MustCompile(`^-?\d*$`)
var emailRx = regexp.MustCompile(`^\S+@\S+$`)
var lettersWithSpacesRx = regexp.MustCompile(`^[- 'a-zA-ZÀ-ÖØ-öø-ÿ]+$`)
var lettersWithNumbersRx = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
var usernameRx = regexp.MustCompile(`^[A-Za-z0-9]+(?:[_-][A-Za-z0-9]+)*$`)
var lettersSpacesAndNumbersRx = regexp.MustCompile(`^[- 'a-zA-ZÀ-ÖØ-öø-ÿ0-9]+$`)
var urlRx = regexp.MustCompile(`^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/)?[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)
var aTrueRx = regexp.MustCompile(`^true$`)

// Regex Confirms that value matches the provided regex
func Regex(field string, errors map[string][]string, rx *regexp.Regexp, value string, message string) bool {
	if !rx.MatchString(value) {
		AddError(errors, field, message)
		return false
	}

	return true
}

// Email Checks that email address is valid
func Email(field string, errors map[string][]string, email string) bool {
	return Regex(field, errors, emailRx, email, fmt.Sprintf("Email address '%s' is invalid", email))
}

// Fail Used for testing, always fails and adds the message
func Fail(field string, errors map[string][]string, message string) bool {
	AddError(errors, field, message)
	return false
}

// NoError used when there's an error.  If there's an error, then it fails
func NoError(field string, errors map[string][]string, err error, message string) bool {
	if err != nil {
		AddError(errors, field, message)
		return false
	}

	return true
}

// False Checks that a value is true
func False(field string, errors map[string][]string, v bool, msg string) bool {
	if v {
		AddError(errors, field, msg)
	}

	return !v
}

// Length Requires the string to be a length between m and n inclusive
func Length(field string, errors map[string][]string, v string, m int, n int) bool {
	var msg string
	if m == n {
		msg = fmt.Sprintf("Must be exactly %d characters long", m)
	} else {
		msg = fmt.Sprintf("Must be between %d and %d characters long", m, n)
	}

	if len(v) < m || len(v) > n {
		AddError(errors, field, msg)
		return false
	}

	return true
}

// MinimumLength Requires the minimum length for a string to be n characters
func MinimumLength(field string, errors map[string][]string, n int, v string) bool {
	if len(v) < n {
		AddError(errors, field, fmt.Sprintf("Must be at least %d characters long", n))
		return false
	}
	return true
}

// StringsNotEqual Verifies that the two provided values are not equal
func StringsNotEqual(field string, errors map[string][]string, a, b, msg string) bool {
	if a == b {
		AddError(errors, field, msg)
		return false
	}

	return true
}

//UUID Verifies that provided string is a UUID
func UUID(field string, errors map[string][]string, id string) bool {
	_, err := uuid.FromString(id)

	if err != nil {
		AddError(errors, field, "Expected UUID for field "+field)
		return false
	}

	return true
}

// Numbers Must be a string containing only numbers
func Numbers(field string, errors map[string][]string, v string) bool {
	return !Regex(field, errors, numberRx, v, "Must be numbers only")
}

// Phone Must be a string containing only numbers
func Phone(field string, errors map[string][]string, v string) bool {
	return Numbers(field, errors, v) && Length(field, errors, v, 7, 13)
}

// Positive Checks that an integer is positive
func Positive(field string, errors map[string][]string, v int) bool {
	if v < 0 {
		AddError(errors, field, "Must be positive")
		return false
	}

	return true
}

// Username Username validation
func Username(field string, errors map[string][]string, v string) bool {
	return !Regex(field, errors, usernameRx, v, "Usernames can only contain letters, numbers, underscores and hyphens, and must not begin or end with an underscore or hyphen")
}

// LettersWithSpaces Must only contain letters and spaces
func LettersWithSpaces(field string, errors map[string][]string, v string) bool {
	return !Regex(field, errors, lettersWithSpacesRx, v, "Must only contain letters and spaces")
}

// LettersWithNumbers Must be a string containing only numbers
func LettersWithNumbers(field string, errors map[string][]string, v string) bool {
	return !Regex(field, errors, lettersWithNumbersRx, v, "Must only contain letters and numbers")
}

// LettersSpacesAndNumbers Must be a string containing only numbers
func LettersSpacesAndNumbers(field string, errors map[string][]string, v string) bool {
	return !Regex(field, errors, lettersSpacesAndNumbersRx, v, "Must only contain letters, spaces and numbers")
}

// URL Checks that URL is valid
func URL(field string, errors map[string][]string, v string) bool {
	return !Regex(field, errors, urlRx, v, "Website URL is invalid")
}

// True Checks that a value is true
func True(field string, errors map[string][]string, v bool, msg string) bool {
	if !v {
		AddError(errors, field, msg)
	}

	return !v
}

// After Checks that a date value is after given date value to be compared with
func After(field string, errors map[string][]string, v civil.Date, dateToCompareWith civil.Date) bool {
	if !v.After(dateToCompareWith) {
		AddError(errors, field, "Date should be after "+dateToCompareWith.String())
		return false
	}

	return true
}

// AfterNow Checks that a date value is after now
func AfterNow(field string, errors map[string][]string, v civil.Date) bool {
	now := civil.DateOf(time.Now().Add((time.Hour * -24)))
	return After(field, errors, v, now)
}

// DOB Checks that a date value is before 18 years from now
func DOB(field string, errors map[string][]string, v civil.Date) bool {
	now := time.Now()
	minAge := civil.Date{
		Day:   (now.Year() - 18),
		Month: now.Month(),
		Year:  now.Day(),
	}
	return After(field, errors, v, minAge)
}

// IN Checks that a value is an item within an array
func IN(field string, errors map[string][]string, v string, array []string) bool {
	for _, element := range array {
		if element == v {
			return true
		}
	}

	AddError(errors, field, "Value not allowed.  Must be one of: "+strings.Join(array, ", "))
	return false
}

// IntsEqual Verifies that the two integers are equal
func IntsEqual(field string, errors map[string][]string, a, b int, msg string) bool {
	if a != b {
		AddError(errors, field, msg)
		return false
	}

	return true
}
