package web

import (
	"net/http"
	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/echo"
)

// devTemplatesHandler allow to easily render a given template from a route of
// the stack. The query parameters are used as data input for the template.
func devTemplatesHandler(c echo.Context) error {
	name := c.Param("name")
	return c.Render(http.StatusOK, name, devData(c))
}

func devData(c echo.Context) echo.Map {
	data := make(echo.Map)
	for k, v := range c.QueryParams() {
		if len(v) > 0 {
			data[k] = v[0]
		}
	}
	if _, ok := data["Domain"]; !ok {
		data["Domain"] = c.Request().Host
	}
	if _, ok := data["ContextName"]; !ok {
		data["ContextName"] = config.DefaultInstanceContext
	}
	return data
}
