package template

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ssawa/LinkLetter/logger"
)

func listTemplates(path string) []string {
	logger.Debug.Printf("Gathering list of templates...")

	templates := []string{}
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f != nil && !f.IsDir() && strings.HasSuffix(path, ".tmpl") {
			templates = append(templates, path)
		}
		return nil
	})

	// We should only use panic sparringly, in this case we're essentially in an unrecoverable
	// state so it's okay
	if err != nil {
		logger.Error.Printf("Error occured while geting templates")
		panic(err)
	}

	return templates
}

// Templator handles the rending of templates for a web application
type Templator struct {
	templates *template.Template
}

// CreateDefaultTemplator creates a templator object with default settings
func CreateDefaultTemplator() *Templator {
	return &Templator{
		templates: template.Must(template.ParseFiles(listTemplates("templates")...)),
	}
}

// RenderTemplate passes the data to the specified template and renders it
func (t Templator) RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := t.templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
