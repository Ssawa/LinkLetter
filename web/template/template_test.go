package template

import (
	"html/template"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFilesWithPaths(t *testing.T) {
	template, err := parseFilesWithPaths("", "test_assets/templates/testfile.tmpl", "test_assets/templates/nested/template.tmpl")
	assert.Nil(t, err)
	assert.NotNil(t, template.Lookup("test_assets/templates/testfile.tmpl"))
	assert.NotNil(t, template.Lookup("test_assets/templates/nested/template.tmpl"))

	template, err = parseFilesWithPaths("test_assets/templates/", "test_assets/templates/testfile.tmpl", "test_assets/templates/nested/template.tmpl")
	assert.Nil(t, err)
	assert.NotNil(t, template.Lookup("testfile.tmpl"))
	assert.NotNil(t, template.Lookup("nested/template.tmpl"))
}

func TestListTemplates(t *testing.T) {
	templates := listTemplates("test_assets/templates")
	assert.Len(t, templates, 2)
	assert.Contains(t, templates, "test_assets/templates/testfile.tmpl")
	assert.Contains(t, templates, "test_assets/templates/nested/template.tmpl")
	assert.NotContains(t, templates, "test_assets/templates/nottemplate.txt")
}

func TestCreateDefaultTemplator(t *testing.T) {
	originalCWD, _ := os.Getwd()
	defer func() { os.Chdir(originalCWD) }()
	os.Chdir("test_assets")

	templator := CreateDefaultTemplator()
	assert.NotNil(t, templator.templates.Lookup("testfile.tmpl"))
	assert.NotNil(t, templator.templates.Lookup("nested/template.tmpl"))
}

func TestRenderTemplate(t *testing.T) {
	templator := Templator{
		templates: template.Must(template.New("test_template").Parse("This is a var: {{ .Value }}")),
	}

	resp := httptest.NewRecorder()
	templator.RenderTemplate(resp, "test_template", struct {
		Value string
	}{Value: "example"})
	assert.Equal(t, "This is a var: example", resp.Body.String())
}
