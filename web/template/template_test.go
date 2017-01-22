package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListTemplates(t *testing.T) {
	templates := listTemplates("test_assets/templates")
	assert.Len(t, templates, 2)
	assert.Contains(t, templates, "templates/testfile.tmpl")
	assert.Contains(t, templates, "templates/nested/template.tmpl")
	assert.NotContains(t, templates, "templates/nottemplate.txt")
}
