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
}

func TestSetupAssetsStatik(t *testing.T) {
	e := echo.New()
	err := SetupAssets(e, "")
	if !assert.NoError(t, err) {
		return
	}

	ts := httptest.NewServer(e)
	defer ts.Close()

}

func TestSetupRoutes(t *testing.T) {
	e := echo.New()
	_, err := SetupRoutes(e)
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
