package router

import (
	"io/fs"
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/web"
)

func LoadTemplatesFromEmbed(engine *gin.Engine) error {
	funcMap := template.FuncMap{
		"escape":           escape,
		"noescape":         noescape,
		"noescapeAttr":     noescapeAttr,
		"oddOrEven":        oddOrEven,
		"visibilityIcon":   visibilityIcon,
		"timestamp":        timestamp,
		"timestampVague":   timestampVague,
		"timestampPrecise": timestampPrecise,
		"emojify":          emojify,
	}

	tmplDir, err := fs.Sub(embed.WebFS, "template")
	if err != nil {
		return err
	}

	tmpl := template.New("").Funcs(funcMap)
	tmpl = template.Must(tmpl.ParseFS(tmplDir, "*.tmpl"))
	
	engine.SetHTMLTemplate(tmpl)
	return nil
}
