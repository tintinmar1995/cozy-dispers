package enclave

import (
	"fmt"
	"os"
	"testing"

	"github.com/cozy/checkup"
	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/prefixer"
)

func TestMain(m *testing.M) {
	config.UseTestFile()

	// First we make sure couchdb is started
	db, err := checkup.HTTPChecker{URL: config.CouchURL().String()}.Check()
	if err != nil || db.Status() != checkup.Healthy {
		fmt.Println("This test need couchdb to run.")
		os.Exit(1)
	}

	prefixerCI = prefixer.TestConceptIndexorPrefixer

	err = couchdb.ResetDB(prefixerCI, "io.cozy.hashconcept")
	if err != nil {
		fmt.Printf("Cant reset db (%s, %s) %s\n", prefixerCI, "io.cozy.hashconcept", err.Error())
		os.Exit(1)
	}

	if err := couchdb.InitGlobalDB(); err != nil {
		fmt.Println("Cant init GlobalDB")
		os.Exit(1)
	}

	res := m.Run()
	os.Exit(res)

}
