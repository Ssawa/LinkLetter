package template

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cj-dimaggio/LinkLetter/logger"
)

// parseFilesWithPaths is almost exactly the same as template.parseFiles with the
// exception that it maintains the relative path to the file rather than parsing out
// the basename. This allows us to have the same filenames in different directories.
// Optionally, you can choose to trim out a base route to clean things up.
// e.x:
//    index.html
//    users/index.html
//    posts/index.html
//    etc...
func parseFilesWithPaths(prefix string, filenames ...string) (*template.Template, error) {
	var t *template.Template
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("html/template: no files named in call to ParseFiles")
	}
	for _, filename := range filenames {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			logger.Error.Printf("Error reading file: %s", err)
			return nil, err
		}
		s := string(b)

		// This is the only real change we're making, where we're replacing:
		//     name := filepath.Base(filename)
		name := strings.TrimPrefix(filename, prefix)

		// First template becomes return value if not already defined,
		// and we use that one for subsequent New calls to associate
		// all the templates together. Also, if this file has the same name
		// as t, this file becomes the contents of t, so
		//  t, err := New(name).Funcs(xxx).ParseFiles(name)
		// works. Otherwise we create a new template associated with t.
		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			logger.Error.Printf("Error parsing template: %s", err)
			return nil, err
		}
	}
	return t, nil
}

// listTemplates lists all the template files in the given path, in no certain order.
func listTemplates(root string) []string {
	logger.Debug.Printf("Gathering list of templates...")

	templates := []string{}
	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
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
		logger.Error.Printf("Error ocurred while geting templates: %s", err)
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
		templates: template.Must(parseFilesWithPaths("templates/", listTemplates("templates")...)),
	}
}

// RenderTemplate passes the data in to the specified template and renders it
func (t Templator) RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := t.templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		logger.Error.Printf("Error rendering template: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
