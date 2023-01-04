package web

import "embed"

//go:embed all:assets
//go:embed all:template
var WebFS embed.FS
