package enclave

import (
	"fmt"
	"os"
	"testing"

	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/prefixer"
	"github.com/cozy/cozy-stack/tests/testutils"
)

func TestMain(m *testing.M) {
	config.UseTestFile()

	// Check is CouchDB is running
	testutils.NeedCouchdb()
	// Run tests over TestDB
	PrefixerCI = prefixer.TestConceptIndexorPrefixer

	// Reinitiate DB
	err := couchdb.ResetDB(PrefixerCI, "io.cozy.hashconcept")
	if err != nil {
		fmt.Printf("Cant reset db (%s, %s) %s\n", PrefixerCI, "io.cozy.hashconcept", err.Error())
		os.Exit(1)
	}
	if err := couchdb.InitGlobalDB(); err != nil {
		fmt.Println("Cant init GlobalDB")
		os.Exit(1)
	}

	res := m.Run()
	os.Exit(res)

}
