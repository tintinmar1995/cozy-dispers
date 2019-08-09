package enclave

import (
	"testing"

	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/stretchr/testify/assert"
)

func TestBuildQuery(t *testing.T) {

	inst := query.Instance{
		Domain:      "prettyname4acozy.mycozy.cloud",
		TokenBearer: "vdf5s4fs2ffse4fc7es5fd",
	}

	localQuery := query.LocalQuery{
		FindRequest: map[string]interface{}{},
	}

	out := buildStackQuery(inst, localQuery)
	assert.Equal(t, inst.Domain, out.Domain)
}
