package pgroup

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstructorNilContext(t *testing.T) {
	var stdout, stderr bytes.Buffer
	p, err := newProcess(nil, "ls -al", withStdOut(&stdout), withStdErr(&stderr))
	assert.Nil(t, err)
	assert.NotNil(t, p)
}

func TestProcess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var stdout, stderr bytes.Buffer
	p, err := newProcess(ctx, "ls -al", withStdOut(&stdout), withStdErr(&stderr))
	assert.Nil(t, err)
	err = p.Run()
	assert.Nil(t, err)
	assert.Equal(t, stderr.String(), "", "stderr should be empty")
	assert.NotEqual(t, stdout.String(), "", "stdout should not be empty")
}
