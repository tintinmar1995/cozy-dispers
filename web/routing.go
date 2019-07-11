//go:generate statik -f -src=../assets -dest=. -externals=../assets/.externals

package web

import (
	"strconv"
	"time"

	build "github.com/cozy/cozy-stack/pkg/config"
	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/metrics"
	"github.com/cozy/cozy-stack/web/errors"
	"github.com/cozy/cozy-stack/web/middlewares"
	"github.com/cozy/cozy-stack/web/query"
	"github.com/cozy/cozy-stack/web/statik"
	"github.com/cozy/cozy-stack/web/status"
	"github.com/cozy/cozy-stack/web/subscribe"
	"github.com/cozy/cozy-stack/web/version"
	"github.com/cozy/echo"
	"github.com/cozy/echo/middleware"
	"github.com/prometheus/client_golang/prometheus"
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

func timersMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			status := strconv.Itoa(c.Response().Status)
			metrics.HTTPTotalDurations.
				WithLabelValues(c.Request().Method, status).
				Observe(v)
		}))
		defer timer.ObserveDuration()
		return next(c)
	}
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

// SetupRoutes returns a new web server that will handle DISPERS' routes
func SetupRoutes(router *echo.Echo) (*echo.Echo, error) {

	if err := SetupAssets(router, config.GetConfig().Assets); err != nil {
		return nil, err
	}

	router.Use(timersMiddleware)

	if !config.GetConfig().CSPDisabled {
		secure := middlewares.Secure(&middlewares.SecureConfig{
			HSTSMaxAge:        hstsMaxAge,
			CSPDefaultSrc:     []middlewares.CSPSource{middlewares.CSPSrcSelf},
			CSPImgSrc:         []middlewares.CSPSource{middlewares.CSPSrcData, middlewares.CSPSrcBlob},
			CSPFrameAncestors: []middlewares.CSPSource{middlewares.CSPSrcNone},
		})
		router.Use(secure)
	}

	router.Use(middlewares.CORS(middlewares.CORSOptions{
		BlackList: []string{},
	}))

	// other non-authentified routes
	{
		query.Routes(router.Group("/dispers"))
		subscribe.Routes(router.Group("/subscribe"))
		status.Routes(router.Group("/status"))
		version.Routes(router.Group("/version"))
	}

	setupRecover(router)
	router.HTTPErrorHandler = errors.ErrorHandler

	main := echo.New()
	main.HideBanner = true
	main.HidePort = true
	main.Renderer = router.Renderer
	main.Any("/*", func(c echo.Context) error {
		router.ServeHTTP(c.Response(), c.Request())
		return nil
	})

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
