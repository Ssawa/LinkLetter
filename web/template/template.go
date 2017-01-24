package template

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ssawa/LinkLetter/logger"
)

// listTemplates lists all the template files in the given path, in no certain order.
func listTemplates(path string) []string {
	logger.Debug.Printf("Gathering list of templates...")

	templates := []string{}
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f != nil && !f.IsDir() && strings.HasSuffix(path, ".tmpl") {
			templates = append(templates, path)
		}
		return nil
	})

	// Go has no real concept of exceptions nor exception handlers. It instead suggests
	// that functions return an error object and if it is not nil, that signifies something
	// went wrong. It's a different, but not necessarily terrible, system, and I do understand
	// the reasoning behind trying to maintain clear program flow. However, Go does provide the
	// "panic" function for cases where something really has gone irreaporably wrong and we need
	// to pull the plug on the application, although it urges the programmer to use this sparingly.
	// Honestly, blocks like the following and ones like it probably *could* be restructured
	// to return an error and not throw a panic. But I make the argument that if no templates can
	// be found than something terrible has happened and there's no chance the application will work
	// correctly, so a panic is justified. Also, adherring to a new paradigm is hard...
	if err != nil {
		logger.Error.Printf("Error ocurred while geting templates")
		panic(err)
	}

	return templates
}

// Templator handles the rending of templates for a web application
type Templator struct {
	templates *template.Template
}

// CreateDefaultTemplator creates a templator object with default settings and caches the parsed
// files for later use.
func CreateDefaultTemplator() *Templator {
	return &Templator{
		templates: template.Must(template.ParseFiles(listTemplates("templates")...)),
	}
}

// RenderTemplate passes the data in to the specified template and renders it
func (t Templator) RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := t.templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
