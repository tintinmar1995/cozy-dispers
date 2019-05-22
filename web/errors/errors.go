package errors

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/jsonapi"
	"github.com/cozy/cozy-stack/pkg/logger"
	"github.com/cozy/cozy-stack/web/middlewares"
	"github.com/sirupsen/logrus"

	"github.com/cozy/echo"
)

// ErrorHandler is the default error handler of our APIs.
func ErrorHandler(err error, c echo.Context) {
	var je *jsonapi.Error
	var ce *couchdb.Error

	res := c.Response()
	req := c.Request()

	var ok bool
	if _, ok = err.(*echo.HTTPError); ok {
		// nothing to do
	} else if os.IsExist(err) {
		je = jsonapi.Conflict(err)
	} else if os.IsNotExist(err) {
		je = jsonapi.NotFound(err)
	} else if ce, ok = err.(*couchdb.Error); ok {
		je = &jsonapi.Error{
			Status: ce.StatusCode,
			Title:  ce.Name,
			Detail: ce.Reason,
		}
	} else if je, ok = err.(*jsonapi.Error); !ok {
		je = &jsonapi.Error{
			Status: http.StatusInternalServerError,
			Title:  "Unqualified error",
			Detail: err.Error(),
		}
	}

	if res.Committed {
		return
	}

	if je != nil {
		if req.Method == http.MethodHead {
			_ = c.NoContent(je.Status)
			return
		}
		_ = jsonapi.DataError(c, je)
		return
	}

	HTMLErrorHandler(err, c)
}

// HTMLErrorHandler is the default fallback error handler for error rendered in
// HTML pages, mainly for users, assets and routes that are not part of our API
// per-se.
func HTMLErrorHandler(err error, c echo.Context) {
	status := http.StatusInternalServerError

	req := c.Request()

	var log *logrus.Entry
	log = logger.WithNamespace("http")
	log.Errorf("%s %s %s", req.Method, req.URL.Path, err)

	var he *echo.HTTPError
	if he, ok := err.(*echo.HTTPError); ok {
		status = he.Code
	} else {
		he = echo.NewHTTPError(status, err)
		he.Inner = err
	}

	var title, value string

	if status >= 500 {
		title = "Error Internal Server Error Title"
		value = "Error Internal Server Error Message"
	} else {
		title = "Error Title"
		value = "Error"
		//value = fmt.Sprintf("%v", he.Message)
	}

	accept := req.Header.Get("Accept")
	acceptHTML := strings.Contains(accept, echo.MIMETextHTML)
	acceptJSON := strings.Contains(accept, echo.MIMEApplicationJSON)
	if req.Method == http.MethodHead {
		err = c.NoContent(status)
	} else if acceptJSON {
		err = c.JSON(status, echo.Map{"error": he.Message})
	} else if acceptHTML {

		err = c.Render(status, "error_new.html", echo.Map{
			"Title":      "Cozy-DISPERS - Erreur",
			"CozyUI":     middlewares.CozyUI("localhost:8080/"),
			"ThemeCSS":   middlewares.ThemeCSS("localhost:8080/"),
			"ErrorTitle": title,
			"Error":      value,
		})
	} else {
		//err = c.String(status, fmt.Sprintf("%v", he.Message))
		err = c.String(status, fmt.Sprintf("Error"))
	}

	if err != nil && log != nil {
		log.Errorf("%s %s %s", req.Method, req.URL.Path, err)
	}
}
