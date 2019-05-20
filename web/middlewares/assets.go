package middlewares

import (
	"html/template"
	"bytes"

	"github.com/cozy/echo"
)

// FuncsMap is a the helper functions used in templates.
// It is filled in web/statik but declared here to avoid circular imports.
var FuncsMap template.FuncMap

var cozyUITemplate *template.Template
var themeTemplate *template.Template
var faviconTemplate *template.Template

// BuildTemplates ensure that the cozy-ui can be injected in templates
func BuildTemplates() {
	cozyUITemplate = template.Must(template.New("cozy-ui").Funcs(FuncsMap).Parse(`` +
		`<link rel="stylesheet" type="text/css">`,
	))
	themeTemplate = template.Must(template.New("theme").Funcs(FuncsMap).Parse(`` +
		`<link rel="stylesheet" type="text/css">`,
	))
	faviconTemplate = template.Must(template.New("favicon").Funcs(FuncsMap).Parse(`
	<link rel="icon">
	<link rel="icon" type="image/png" sizes="16x16">
	<link rel="icon" type="image/png" sizes="32x32">
	<link rel="apple-touch-icon" sizes="180x180"/>
		`))
}


// CozyUI returns an HTML template to insert the Cozy-UI assets.
func CozyUI() template.HTML {
	buf := new(bytes.Buffer)
	err := cozyUITemplate.Execute(buf, echo.Map{
	})
	if err != nil {
		panic(err)
	}
	return template.HTML(buf.String())
}

// ThemeCSS returns an HTML template for inserting the HTML tag for the custom
// CSS theme
func ThemeCSS() template.HTML {
	buf := new(bytes.Buffer)
	err := themeTemplate.Execute(buf, echo.Map{
	})
	if err != nil {
		panic(err)
	}
	return template.HTML(buf.String())
}

// Favicon returns a helper to insert the favicons in an HTML template.
func Favicon() template.HTML {
	buf := new(bytes.Buffer)
	err := faviconTemplate.Execute(buf, echo.Map{
	})
	if err != nil {
		panic(err)
	}
	return template.HTML(buf.String())
}
