package render

import (
	"booking/models"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddDefaultData(t *testing.T) {
	td := models.TemplateData{}

	r, err := getSession()
	assert.NoError(t, err)

	session.Put(r.Context(), "flash", "123")

	res := AddDefaultData(&td, r)
	assert.Equal(t, "123", res.Flash)
}

func TestRenderTemplate(t *testing.T) {
	pathToTemplate = "./../templates"
	tc, err := CreateTemplateCache()
	assert.NoError(t, err)

	app.TemplateCache = tc

	r, err := getSession()
	assert.NoError(t, err)
	w := &myWriter{}

	err = RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
	assert.NoError(t, err)

	err = RenderTemplate(w, r, "non-existant.page.tmpl", &models.TemplateData{})
	assert.Error(t, err)
}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplate = "./../templates"
	_, err := CreateTemplateCache()
	assert.NoError(t, err)
}

func getSession() (*http.Request, error) {
	r, err := http.NewRequest(http.MethodGet, "/some-url", nil)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))
	r = r.WithContext(ctx)
	return r, nil
}
