package middlewares

import (
	"html/template"
)

// FuncsMap is a the helper functions used in templates.
// It is filled in web/statik but declared here to avoid circular imports.
var FuncsMap template.FuncMap
