package testutils

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/cozy/checkup"
	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/utils"
	"github.com/cozy/echo"
)

// This flag avoid starting the stack twice.
var stackStarted bool

// Fatal prints a message and immediately exit the process
func Fatal(msg ...interface{}) {
	fmt.Println(msg...)
	os.Exit(1)
}

// NeedCouchdb kill the process if there is no couchdb running
func NeedCouchdb() {
	db, err := checkup.HTTPChecker{URL: config.CouchURL().String()}.Check()
	if err != nil || db.Status() != checkup.Healthy {
		Fatal("This test need couchdb to run.")
	}
}

// NeedOtherDispersServer kill the process if there is no other Server Dispers running
func NeedOtherDispersServer(dispersURL url.URL) {
	dispersURL.Path = "version"
	dis, err := checkup.HTTPChecker{URL: dispersURL.String()}.Check()
	if err != nil || dis.Status() != checkup.Healthy {
		Fatal("This test need another server DISPERS on port 8118 to run.")
	}
}

// TestSetup is a wrapper around a testing.M which handles
// setting up instance, client, VFSContext, testserver
// and cleaning up after itself
type TestSetup struct {
	testingM *testing.M
	name     string
	host     string
	ts       *httptest.Server
	cleanup  func()
}

// NewSetup returns a new TestSetup
// name is used to prevent bug when tests are run in parallel
func NewSetup(testingM *testing.M, name string) *TestSetup {
	setup := TestSetup{
		name:     name,
		testingM: testingM,
		host:     name + "_" + utils.RandomString(10) + ".cozy.local",
		cleanup:  func() {},
	}
	return &setup
}

// CleanupAndDie cleanup the TestSetup, prints a message and 	close the process
func (c *TestSetup) CleanupAndDie(msg ...interface{}) {
	c.cleanup()
	Fatal(msg...)
}

// Cleanup cleanup the TestSetup
func (c *TestSetup) Cleanup() {
	c.cleanup()
}

// AddCleanup adds a function to be run when the test is finished.
func (c *TestSetup) AddCleanup(f func() error) {
	next := c.cleanup
	c.cleanup = func() {
		err := f()
		if err != nil {
			fmt.Println("Error while cleanup", err)
		}
		next()
	}
}

// GetTmpDirectory creates a temporary directory
// The directory will be removed on container cleanup
func (c *TestSetup) GetTmpDirectory() string {
	tempdir, err := ioutil.TempDir("", "cozy-stack")
	if err != nil {
		c.CleanupAndDie("Could not create temporary directory.", err)
	}
	c.AddCleanup(func() error { return os.RemoveAll(tempdir) })
	return tempdir
}

// stupidRenderer is a renderer for echo that does nothing.
// It is used just to avoid the error "Renderer not registered" for rendering
// error pages.
type stupidRenderer struct{}

func (sr *stupidRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return nil
}

// Run runs the underlying testing.M and cleanup
func (c *TestSetup) Run() int {
	value := c.testingM.Run()
	c.cleanup()
	return value
}

// CookieJar is a http.CookieJar which always returns all cookies.
// NOTE golang stdlib uses cookies for the URL (ie the testserver),
// not for the host (ie the instance), so we do it manually
type CookieJar struct {
	Jar *cookiejar.Jar
	URL *url.URL
}

// Cookies implements http.CookieJar interface
func (j *CookieJar) Cookies(u *url.URL) (cookies []*http.Cookie) {
	return j.Jar.Cookies(j.URL)
}

// SetCookies implements http.CookieJar interface
func (j *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.Jar.SetCookies(j.URL, cookies)
}
