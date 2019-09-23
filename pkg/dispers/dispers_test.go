package enclave

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/dispers/network"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/cozy/cozy-stack/pkg/prefixer"
	"github.com/cozy/cozy-stack/tests/testutils"
)

var testDispersURL = url.URL{Host: "cozy.tools:8008", Scheme: "http"}

func TestMain(m *testing.M) {
	config.UseTestFile()

	// Check is CouchDB is running
	testutils.NeedCouchdb()
	// Run tests over TestDB
	PrefixerC = prefixer.TestConductorPrefixer
	PrefixerCI = prefixer.TestConceptIndexorPrefixer
	query.PrefixerC = prefixer.TestConductorPrefixer
	query.PrefixerT = prefixer.TestTargetPrefixer

	// Reinitiate DB
	err := couchdb.ResetDB(PrefixerCI, "io.cozy.hashconcept")
	if err != nil {
		fmt.Printf("Cant reset db (%s, %s) %s\n", PrefixerCI, "io.cozy.hashconcept", err.Error())
		os.Exit(1)
	}
	err = couchdb.ResetDB(PrefixerC, "io.cozy.query")
	if err != nil {
		fmt.Printf("Cant reset db (%s, %s) %s\n", PrefixerC, "io.cozy.query", err.Error())
		os.Exit(1)
	}
	err = couchdb.ResetDB(PrefixerC, "io.cozy.instances")
	if err != nil {
		fmt.Printf("Cant reset db (%s, %s) %s\n", PrefixerC, "io.cozy.instances", err.Error())
		os.Exit(1)
	}
	err = couchdb.ResetDB(PrefixerC, "io.cozy.async")
	if err != nil {
		fmt.Printf("Cant reset db (%s, %s) %s\n", PrefixerC, "io.cozy.async", err.Error())
		os.Exit(1)
	}
	couchdb.InitGlobalDB()

	testutils.NeedOtherDispersServer(testDispersURL)

	// Hosts is used for conductor_test
	ConductorURL = testDispersURL
	network.Hosts = []url.URL{testDispersURL}
	res := m.Run()
	os.Exit(res)
}
