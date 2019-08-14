package metadata

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/cozy/checkup"
	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/prefixer"
	"github.com/stretchr/testify/assert"
)

func TestExecutionMetadata(t *testing.T) {

	_, err := RetrieveExecutionMetadata("mean")
	assert.Error(t, err)

	// Let's Create A ExecutionMetadata
	meta, err := NewExecutionMetadata("subscribe", "Cozy-DISPERS", url.URL{Scheme: "http"})
	assert.NoError(t, err)

	// Let's Create Task For The ExecutionMetadata
	task := NewTaskMetadata()
	err = meta.HandleError("Hope It Will Not Crash", task, err)
	assert.NoError(t, err)
	err = meta.EndExecution(errors.New("Oh nooooo!"))
	assert.NoError(t, err)

	// Let's Create A ExecutionMetadata
	meta, err = NewExecutionMetadata("query", "mean", url.URL{Scheme: "http"})
	assert.NoError(t, err)

	meta2, err := RetrieveExecutionMetadata("mean")
	assert.NoError(t, err)
	assert.Equal(t, meta.ID(), meta2.ID())
	assert.Equal(t, meta.Rev(), meta2.Rev())

	err = meta2.EndExecution(nil)
	assert.NoError(t, err)
}

func TestMain(m *testing.M) {
	config.UseTestFile()

	prefixC = prefixer.TestConductorPrefixer

	// First we make sure couchdb is started
	db, err := checkup.HTTPChecker{URL: config.CouchURL().String()}.Check()
	if err != nil || db.Status() != checkup.Healthy {
		fmt.Println("This test need couchdb to run.")
		os.Exit(1)
	}

	err = couchdb.ResetDB(prefixC, "io.cozy.execution.metadata")
	if err != nil {
		fmt.Printf("Cant reset db (%s, %s) %s\n", prefixC, "io.cozy.execution.metadata", err.Error())
		os.Exit(1)
	}

	couchdb.InitGlobalDB()

	res := m.Run()
	os.Exit(res)

}
