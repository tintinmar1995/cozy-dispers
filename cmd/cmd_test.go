package cmd

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"github.com/cozy/checkup"
	"github.com/cozy/cozy-stack/client"
	"github.com/cozy/cozy-stack/model/stack"
	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/web"
	"github.com/cozy/echo"
)

var testClient *client.Client

func TestMain(m *testing.M) {
	config.UseTestFile()

	db, err := checkup.HTTPChecker{URL: config.CouchURL().String()}.Check()
	if err != nil || db.Status() != checkup.Healthy {
		fmt.Println("This test need couchdb to run.")
		os.Exit(1)
	}

	_, err = stack.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tempdir, err := ioutil.TempDir("", "cozy-stack")
	if err != nil {
		fmt.Println("Could not create temporary directory.")
		os.Exit(1)
	}

	config.GetConfig().Fs.URL = &url.URL{
		Scheme: "file",
		Host:   "localhost",
		Path:   tempdir,
	}
	server := echo.New()
	err = web.SetupRoutes(server)
	if err != nil {
		fmt.Println("Could not start server", err)
		os.Exit(1)
	}

	res := m.Run()
	os.RemoveAll(tempdir)
	os.Exit(res)
}
