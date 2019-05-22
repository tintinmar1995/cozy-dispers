package middlewares

import (
	"bytes"
	"html/template"

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
		`<link rel="stylesheet" type="text/css" href="{{asset .Host "/css/cozy-ui.min.css"}}">`,
	))
	themeTemplate = template.Must(template.New("theme").Funcs(FuncsMap).Parse(`` +
		`<link rel="stylesheet" type="text/css" href="{{asset .Host "/styles/theme.css"}}">`,
	))
	faviconTemplate = template.Must(template.New("favicon").Funcs(FuncsMap).Parse(`
	<link rel="icon" href="{{asset .Host "/favicon.ico"}}">
	<link rel="icon" type="image/png" href="{{asset .Host "/favicon-16x16.png"}}" sizes="16x16">
	<link rel="icon" type="image/png" href="{{asset .Host "/favicon-32x32.png"}}" sizes="32x32">
	<link rel="apple-touch-icon" sizes="180x180" href="{{asset .Host "/apple-touch-icon.png"}}"/>
		`))
}

// CozyUI returns an HTML template to insert the Cozy-UI assets.
func CozyUI(host string) template.HTML {
	buf := new(bytes.Buffer)
	err := cozyUITemplate.Execute(buf, echo.Map{
		"Host": host,
	})
	if err != nil {
		panic(err)
	}
	return template.HTML(buf.String())
}

// ThemeCSS returns an HTML template for inserting the HTML tag for the custom
// CSS theme
func ThemeCSS(host string) template.HTML {
	buf := new(bytes.Buffer)
	err := themeTemplate.Execute(buf, echo.Map{
		"Host": host,
	})
	if err != nil {
		panic(err)
	}
	return template.HTML(buf.String())
}

// Favicon returns a helper to insert the favicons in an HTML template.
func Favicon(host string) template.HTML {
	buf := new(bytes.Buffer)
	err := faviconTemplate.Execute(buf, echo.Map{
		"Host": host,
	})
	if err != nil {
		panic(err)
	}
	return template.HTML(buf.String())
}
