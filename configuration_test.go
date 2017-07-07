package pgroup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigurationMethodsWhenNil(t *testing.T) {
	var c *configuration
	assert.Nil(t, c)
	assert.NotNil(t, c.stdout())
	assert.NotNil(t, c.stderr())
}
