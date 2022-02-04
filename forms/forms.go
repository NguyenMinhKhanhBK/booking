package form

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

type Form struct {
	url.Values
	Errors errors
}

func New(data url.Values) *Form {
	return &Form{
		Values: data,
		Errors: make(errors),
	}
}

func (f *Form) Require(fields ...string) {
	for _, field := range fields {
		val := f.Get(field)
		if strings.TrimSpace(val) == "" {
			f.Errors.Add(field, "This field cannot be empty")
		}
	}
}

func (f *Form) Has(field string) bool {
	if f.Get(field) == "" {
		return false
	}

	return true
}

func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

func (f *Form) MinLength(field string, length int) bool {
	x := f.Get(field)
	if len(x) < length {
		errMsg := fmt.Sprintf("This field must be at least %d characters long", length)
		f.Errors.Add(field, errMsg)
		return false
	}
	return true
}

func (f *Form) IsEmail(field string) bool {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
		return false
	}
	return true
}
