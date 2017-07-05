package pgroup

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroup(t *testing.T) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, err := New(ctx, WithStdOut(os.Stdout), WithStdErr(os.Stdout))
	assert.Nil(t, err)

	err = g.NewProcess("dir1", "ls -al")
	assert.Nil(t, err)

	err = g.NewProcess("dir2", "ls -al")
	assert.Nil(t, err)

	err = g.Run()
	assert.Nil(t, err)

}

func TestSetEnv(t *testing.T) {
	var err error
	var stdout bytes.Buffer
	envKey := "FOO=BAR"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, err := New(ctx, WithStdOut(&stdout), WithEnv([]string{envKey}))
	assert.Nil(t, err)

	err = g.NewProcess("env1", "env")
	assert.Nil(t, err)

	err = g.Run()
	assert.Nil(t, err)
	s := stdout.String()
	if !strings.Contains(s, envKey) {
		t.Fatalf("Env key %s is missing from the commands environment.", envKey)
	}
}
