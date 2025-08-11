package resources

import "embed"

//go:embed web/index.html
var IndexHtmlTemplate string

//go:embed web
var StaticHtmlFS embed.FS
