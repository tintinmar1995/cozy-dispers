package enclave

import (
	"testing"

	"github.com/cozy/cozy-stack/pkg/dispers/dispers"
	"github.com/stretchr/testify/assert"
)

func TestBuildQuery(t *testing.T) {

	inst := dispers.Instance{
		Domain: "prettyname4acozy.mycozy.cloud",
		Token:  dispers.Token{TokenBearer: "vdf5s4fs2ffse4fc7es5fd"},
	}

	localQuery := dispers.LocalQuery{
		FindRequest: map[string]interface{}{},
	}

	in := dispers.InputT{
		LocalQuery: localQuery,
	}
	out := buildQuery(inst, in.LocalQuery)
	assert.Equal(t, inst.Domain, out.Domain)
}
