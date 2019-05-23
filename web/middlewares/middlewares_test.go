package middlewares_test

import (
	"testing"

	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/web/middlewares"
	"github.com/stretchr/testify/assert"
)

func TestSplitHost(t *testing.T) {
	config.UseTestFile()

	host, app, siblings := middlewares.SplitHost("localhost")
	assert.Equal(t, "localhost", host)
	assert.Equal(t, "", app)
	assert.Equal(t, "", siblings)

	host, app, siblings = middlewares.SplitHost("joe.example.net")
	assert.Equal(t, "joe.example.net", host)
	assert.Equal(t, "", app)
	assert.Equal(t, "", siblings)
}
