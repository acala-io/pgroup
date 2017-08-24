package pgroup

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcessMethodsWhenNil(t *testing.T) {
	var p *process
	assert.Nil(t, p)
	assert.Equal(t, ErrNotConfigured, p.Run())
	assert.Equal(t, ErrNotConfigured, p.Signal(syscall.SIGHUP))
	assert.Equal(t, ErrNotConfigured, p.AddEnv("foo", "bar"))
}

func TestConfiguration(t *testing.T) {
	var stdout, stderr bytes.Buffer
	p, err := newProcess(nil, "ls -al", withStdOut(&stdout), withStdErr(&stderr), withEnv([]string{"FOO=BAR"}))
	assert.Nil(t, err)
	assert.NotNil(t, p)
	// Running configuration again should return errors
	assert.NotNil(t, withStdOut(&stdout)(p))
	assert.NotNil(t, withStdErr(&stderr)(p))
	assert.NotNil(t, withEnv([]string{"FOO=BAR"})(p))
}

func TestOptionFuncErrors(t *testing.T) {
	myErr := errors.New("My Error")
	p, err := newProcess(nil, "ls -al", func(p *process) error { return myErr })
	assert.Nil(t, p)
	assert.NotNil(t, err)
	assert.Equal(t, err, myErr)
}

func TestNoConfiguration(t *testing.T) {
	p, err := newProcess(nil, "ls -al")
	assert.Nil(t, err)
	assert.NotNil(t, p)
	err = p.Run()
	assert.Nil(t, err)
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

func TestAddEnv(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var stdout, stderr bytes.Buffer
	p, err := newProcess(ctx, "printenv BAR", withStdOut(&stdout))
	assert.Nil(t, err)
	p.AddEnv("BAR", "2")
	err = p.Run()
	assert.Nil(t, err)
	assert.Equal(t, stderr.String(), "", "stderr should be empty")
	assert.Equal(t, strings.TrimSpace(stdout.String()), "2", "stdout should not be empty")
}

func TestEnv(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var stdout, stderr bytes.Buffer
	os.Setenv("BAR", "1")
	p, err := newProcess(ctx, "printenv BAR", withStdOut(&stdout))
	assert.Nil(t, err)
	err = p.Run()
	assert.Nil(t, err)
	assert.Equal(t, stderr.String(), "", "stderr should be empty")
	assert.Equal(t, strings.TrimSpace(stdout.String()), "1", "stdout should not be empty")
}

func TestCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()

	var stdout, stderr bytes.Buffer
	p, err := newProcess(ctx, "./scripts/cancel.sh", withStdOut(io.MultiWriter(&stdout, os.Stdout)), withStdErr(io.MultiWriter(&stderr, os.Stderr)))
	if err != nil {
		t.Log(err.Error())
	}
	assert.Nil(t, err)
	err = p.Run()
	if err != nil {
		t.Log(err.Error())
	}

	assert.Nil(t, err)
	assert.NotEqual(t, stdout.String(), "", "stdout should not be empty")
}
