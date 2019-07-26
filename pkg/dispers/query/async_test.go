package query

import (
	"fmt"
	"os"
	"testing"

	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/prefixer"
	"github.com/cozy/cozy-stack/tests/testutils"
	"github.com/stretchr/testify/assert"
)

func TestAsync(t *testing.T) {

	state, err := FetchAsyncState("testquery", 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, Waiting, state)

	doc, err := NewAsyncTask("testquery", 0, 0, AsyncAggregation)
	assert.NoError(t, err)

	state, err = FetchAsyncState("testquery", 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, Finished, state)

	doc.Data = map[string]interface{}{"hey": "you"}
	couchdb.UpdateDoc(PrefixerC, &doc)
	data, err := FetchAsyncData("testquery", 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"hey": "you"}, data)

}

func TestMain(m *testing.M) {
	config.UseTestFile()

	// Check is CouchDB is running
	testutils.NeedCouchdb()
	// Run tests over TestDB
	PrefixerC = prefixer.TestConductorPrefixer

	// Reinitiate DB
	err := couchdb.ResetDB(PrefixerC, "io.cozy.async")
	if err != nil {
		fmt.Printf("Cant reset db (%s, %s) %s\n", PrefixerC, "io.cozy.instances", err.Error())
		os.Exit(1)
	}
	couchdb.InitGlobalDB()

	// Hosts is used for conductor_test
	res := m.Run()
	os.Exit(res)
}
