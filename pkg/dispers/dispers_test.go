package enclave

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/dispers/network"
	"github.com/cozy/cozy-stack/pkg/prefixer"
	"github.com/cozy/cozy-stack/tests/testutils"
)

var dispersURL = url.URL{Host: "localhost:8118", Scheme: "http"}

func TestMain(m *testing.M) {
	config.UseTestFile()

	// Check is CouchDB is running
	testutils.NeedCouchdb()
	// Run tests over TestDB
	PrefixerC = prefixer.TestConductorPrefixer
	PrefixerCI = prefixer.TestConceptIndexorPrefixer

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
	couchdb.InitGlobalDB()

	testutils.NeedOtherDispersServer(dispersURL)

	// Hosts is used for conductor_test
	hosts = []url.URL{dispersURL}
	network.Hosts = hosts
	res := m.Run()
	os.Exit(res)
}
