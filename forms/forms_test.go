package form

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/url", nil)
	f := New(r.PostForm)

	isValid := f.Valid()
	assert.True(t, isValid)
}

func TestFrom_Require(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/url", nil)
	f := New(r.PostForm)

	f.Require("a", "b", "c")
	// Form should be invalid
	assert.False(t, f.Valid())

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	r = httptest.NewRequest(http.MethodPost, "/url", nil)
	r.PostForm = postedData
	f = New(r.PostForm)
	f.Require("a", "b", "c")
	assert.True(t, f.Valid())
}

func TestForm_Has(t *testing.T) {
	postedData := url.Values{}
	postedData.Add("a", "a")

	f := New(postedData)
	assert.True(t, f.Has("a"))
	assert.False(t, f.Has("b"))
}

func TestForm_MinLength(t *testing.T) {
	postedData := url.Values{}
	postedData.Add("a", "123456")
	f := New(postedData)
	assert.True(t, f.MinLength("a", 6))
	assert.True(t, f.Valid())

	postedData = url.Values{}
	postedData.Add("b", "123456")

	f = New(postedData)
	assert.False(t, f.MinLength("b", 10))
	assert.False(t, f.Valid())
}

func TestForm_IsEmail(t *testing.T) {
	f := New(nil)
	assert.False(t, f.IsEmail("email"))
	assert.False(t, f.Valid())

	postedData := url.Values{}
	postedData.Add("email", "abc@abc")
	f = New(postedData)
	assert.False(t, f.IsEmail("email"))
	assert.False(t, f.Valid())

	postedData = url.Values{}
	postedData.Add("email", "abc@abc.com")
	f = New(postedData)
	assert.True(t, f.IsEmail("email"))
	assert.True(t, f.Valid())
}

func TestForm_GetError(t *testing.T) {
	f := New(nil)
	assert.False(t, f.IsEmail("email"))
	// should return non-empty error message
	assert.NotEqual(t, "", f.Errors.Get("email"))

	postedData := url.Values{}
	postedData.Add("email", "abc@abc.com")
	f = New(postedData)
	assert.True(t, f.IsEmail("email"))
	// should return no error
	assert.Equal(t, "", f.Errors.Get("email"))
}
