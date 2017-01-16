package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ssawa/LinkLetter/logger"
)

func listTemplates() []string {
	logger.Debug.Printf("Gathering list of templates...")

	templates := []string{}
	err := filepath.Walk("templates", func(path string, f os.FileInfo, err error) error {
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

func (server Server) renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := server.templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
