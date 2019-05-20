//go:generate statik -f -src=../assets -dest=. -externals=../assets/.externals

package web

import (
	//"strconv"
	"time"

	build "github.com/cozy/cozy-stack/pkg/config"
	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/metrics"
	"github.com/cozy/cozy-stack/web/dispers"
	"github.com/cozy/cozy-stack/web/errors"
	"github.com/cozy/cozy-stack/web/middlewares"
	"github.com/cozy/cozy-stack/web/statik"
	"github.com/cozy/cozy-stack/web/version"
	"github.com/cozy/echo"
	"github.com/cozy/echo/middleware"
	//"github.com/prometheus/client_golang/prometheus"
)

const (
	// cspScriptSrcWhitelist is a whitelist for default allowed domains in CSP.
	cspScriptSrcWhitelist = "https://piwik.cozycloud.cc https://matomo.cozycloud.cc"

	// cspImgSrcWhitelist is a whitelist of images domains that are allowed in
	// CSP.
	cspImgSrcWhitelist = "https://piwik.cozycloud.cc https://matomo.cozycloud.cc " +
		"https://*.tile.openstreetmap.org https://*.tile.osm.org " +
		"https://*.tiles.mapbox.com https://api.mapbox.com"

	// cspFrameSrcWhiteList is a whitelist of custom protocols that are allowed
	// in the CSP. We are using iframes on these custom protocols to open
	// deeplinks to them and have a fallback if the mobile apps are not
	// available.
	cspFrameSrcWhiteList = "cozydrive: cozybanks:"
)

var hstsMaxAge = 365 * 24 * time.Hour // 1 year

// SetupAssets add assets routing and handling to the given router. It also
// adds a Renderer to render templates.
func SetupAssets(router *echo.Echo, assetsPath string) (err error) {
	var r statik.AssetRenderer
	if assetsPath != "" {
		r, err = statik.NewDirRenderer(assetsPath)
	} else {
		r, err = statik.NewRenderer()
	}
	if err != nil {
		return err
	}
	middlewares.BuildTemplates()

	cacheControl := middlewares.CacheControl(middlewares.CacheOptions{
		MaxAge: 24 * time.Hour,
	})

	router.Renderer = r
	router.HEAD("/assets/*", echo.WrapHandler(r))
	router.GET("/assets/*", echo.WrapHandler(r))
	router.GET("/favicon.ico", echo.WrapHandler(r), cacheControl)
	router.GET("/robots.txt", echo.WrapHandler(r), cacheControl)
	router.GET("/security.txt", echo.WrapHandler(r), cacheControl)
	return nil
}


// SetupAdminRoutes sets the routing for the administration HTTP endpoints
func SetupAdminRoutes(router *echo.Echo) error {
	var mws []echo.MiddlewareFunc
	if build.IsDevRelease() {
		mws = append(mws, middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "time=${time_rfc3339}\tstatus=${status}\tmethod=${method}\thost=${host}\turi=${uri}\tbytes_out=${bytes_out}\n",
		}))
	} else {
		mws = append(mws, middlewares.BasicAuth(config.GetConfig().AdminSecretFileName))
	}

	version.Routes(router.Group("/version", mws...))
	metrics.Routes(router.Group("/metrics", mws...))

	setupRecover(router)

	router.HTTPErrorHandler = errors.ErrorHandler
	return nil
}

// SetupMajorRoutes (previous CreateSubdomainProxy) returns a new web server that will handle that apps
// proxy routing if the host of the request match an application, and route to
// the given router otherwise.
func SetupMajorRoutes(router *echo.Echo /*, appsHandler echo.HandlerFunc*/) (*echo.Echo, error) {
	if err := SetupAssets(router, config.GetConfig().Assets); err != nil {
		return nil, err
	}

	main := echo.New()
	main.HideBanner = true
	main.HidePort = true
	main.Renderer = router.Renderer

	dispers.Routes(main.Group("/dispers"))
	setupRecover(router)

	main.HTTPErrorHandler = errors.HTMLErrorHandler
	return main, nil
}

// setupRecover sets a recovering strategy of panics happening in handlers
func setupRecover(router *echo.Echo) {
	if !build.IsDevRelease() {
		recoverMiddleware := middlewares.RecoverWithConfig(middlewares.RecoverConfig{
			StackSize: 10 << 10, // 10KB
		})
		router.Use(recoverMiddleware)
	}
}
