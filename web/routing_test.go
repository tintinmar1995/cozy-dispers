package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/tests/testutils"
	"github.com/cozy/echo"
	"github.com/stretchr/testify/assert"
)

var domain string

func TestSetupAssets(t *testing.T) {
	e := echo.New()
	err := SetupAssets(e, "../assets")
	if !assert.NoError(t, err) {
		return
	}

	ts := httptest.NewServer(e)
	defer ts.Close()

	{
		res, err := http.Get(ts.URL + "/assets/images/cozy.svg")
		assert.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, 200, res.StatusCode)
	}

	{
		res, err := http.Head(ts.URL + "/assets/images/cozy.svg")
		assert.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, 200, res.StatusCode)
	}
}

func TestSetupAssetsStatik(t *testing.T) {
	e := echo.New()
	err := SetupAssets(e, "")
	if !assert.NoError(t, err) {
		return
	}

	ts := httptest.NewServer(e)
	defer ts.Close()

	{
		res, err := http.Get(ts.URL + "/assets/images/cozy.svg")
		assert.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, 200, res.StatusCode)
		assert.NotContains(t, res.Header.Get("Cache-Control"), "max-age=")
	}

	{
		res, err := http.Get(ts.URL + "/assets/images/cozy.badbeefbadbeef.svg")
		assert.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, 200, res.StatusCode)
		assert.Contains(t, res.Header.Get("Cache-Control"), "max-age=")
	}

	{
		res, err := http.Get(ts.URL + "/assets/images/cozy.immutable.svg")
		assert.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, 200, res.StatusCode)
		assert.Contains(t, res.Header.Get("Cache-Control"), "max-age=")
	}

	{
		res, err := http.Head(ts.URL + "/assets/images/cozy.svg")
		assert.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, 200, res.StatusCode)
	}

	{
		res, err := http.Head(ts.URL + "/assets/images/cozy.badbeefbadbeef.svg")
		assert.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, 200, res.StatusCode)
		assert.Contains(t, res.Header.Get("Cache-Control"), "max-age=")
	}
}

func TestSetupRoutes(t *testing.T) {
	e := echo.New()
	err := SetupRoutes(e)
	if !assert.NoError(t, err) {
		return
	}

	ts := httptest.NewServer(e)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/version")
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode)
}

func TestMain(m *testing.M) {
	config.UseTestFile()
	config.GetConfig().Assets = "../assets"
	testutils.NeedCouchdb()
	setup := testutils.NewSetup(m, "routing_test")
	os.Exit(setup.Run())
}
