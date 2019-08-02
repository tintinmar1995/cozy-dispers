package network

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleError(t *testing.T) {

	act := &ExternalActor{
		Method: "POST",
		URL: url.URL{
			Scheme: "http",
			Host:   "cozy.io",
			Path:   "dispers",
		},
	}

	act.Outstr = "{\"error\": \"code=403, message=Forbidden\"}"
	act.Out = []byte(act.Outstr)

	act.handleError()
	assert.Equal(t, "403", act.Status)

}
