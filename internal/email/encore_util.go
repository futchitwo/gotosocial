package email

import (
	"io/fs"
	"html/template"

 	"github.com/superseriousbusiness/gotosocial/web"
)

func loadTemplatesFromEmbed() (*template.Template, error) {
	tmplDir, err := fs.Sub(embed.WebFS, "template")
 	if err != nil {
 		return err
 	}

 	tmpl := template.New("").Funcs(funcMap)
	// look for all templates that start with 'email_'
 	return tmpl.ParseFS(tmplDir, "email_*")
}
