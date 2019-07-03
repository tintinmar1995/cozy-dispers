package enclave

import (
	"testing"

	"github.com/cozy/cozy-stack/pkg/dispers/dispers"
	"github.com/stretchr/testify/assert"
)

func TestBuildQuery(t *testing.T) {

	listOfInstances := []dispers.Instance{}

	inst := dispers.Instance{
		Host:   "mycozy.cloud",
		Domain: "prettyname4acozy",
		Token:  dispers.Token{TokenBearer: "vdf5s4fs2ffse4fc7es5fd"},
	}

	localQuery := dispers.LocalQuery{
		FindRequest: map[string]interface{}{},
	}

	in := dispers.InputT{
		LocalQuery: localQuery,
		Targets:    listOfInstances,
	}
	out := buildQuery(inst, in.LocalQuery)
	assert.Equal(t, inst.Domain, out.Domain)
}
