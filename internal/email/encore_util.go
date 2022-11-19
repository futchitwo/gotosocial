package email

import (
	"io/fs"
	"html/template"

 	"github.com/superseriousbusiness/gotosocial/web"
)

func loadTemplatesFromEmbed() (*template.Template, error) {
	tmplDir, err := fs.Sub(web.WebFS, "template")
 	if err != nil {
 		return err
 	}

 	tmpl := template.New("")
	// look for all templates that start with 'email_'
 	err := tmpl.ParseFS(tmplDir, "email_*")
	return tmpl, err
}
