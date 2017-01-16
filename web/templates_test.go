package web

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListTemplates(t *testing.T) {
	originalCWD, _ := os.Getwd()
	defer func() { os.Chdir(originalCWD) }()

	os.Chdir("test_assets")

	templates := listTemplates()
	assert.Len(t, templates, 2)
	assert.Contains(t, templates, "templates/testfile.tmpl")
	assert.Contains(t, templates, "templates/nested/template.tmpl")
	assert.NotContains(t, templates, "templates/nottemplate.txt")
}
