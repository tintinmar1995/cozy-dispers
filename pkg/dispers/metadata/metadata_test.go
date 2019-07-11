package dispers

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
	task, err := NewTaskMetadata()
	assert.NoError(t, err)
	err = meta.SetTask("Hope It Will Not Crash", task)
	assert.NoError(t, err)

	err = meta.EndExecution(errors.New("Oh nooooo!"))
	assert.NoError(t, err)

	// Let's Create A ExecutionMetadata
	meta, err = NewExecutionMetadata("query", "mean", url.URL{Scheme: "http"})
	assert.NoError(t, err)

	meta2, err := RetrieveExecutionMetadata("mean")
	assert.Equal(t, meta.ID(), meta2.ID())
	assert.Equal(t, meta.Rev(), meta2.Rev())
	assert.NoError(t, err)

	// Let's Create Task For The ExecutionMetadata
	task, err = NewTaskMetadata()
	assert.NoError(t, err)
	err = meta2.SetTask("Concept Indexor", task)
	assert.NoError(t, err)
	assert.Equal(t, true, task.Start.Equal(meta2.Tasks["Concept Indexor"].Start))

	// Let's Create Task For The ExecutionMetadata
	task, err = NewTaskMetadata()
	assert.NoError(t, err)
	err = meta2.SetTask("Target Finder", task)
	assert.NoError(t, err)
	assert.Equal(t, true, task.Start.Equal(meta2.Tasks["Target Finder"].Start))

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

	couchdb.InitGlobalTestDB()

	res := m.Run()
	os.Exit(res)

}
