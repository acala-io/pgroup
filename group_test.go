package pgroup

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type fakeProcess struct {
	mock.Mock
}

func (fp *fakeProcess) Run() error {
	args := fp.Called()
	return args.Error(0)
}

func (fp *fakeProcess) Signal(s syscall.Signal) error {
	args := fp.Called(s)
	return args.Error(0)
}

func (fp *fakeProcess) AddEnv(key, value string) error {
	return nil
}

func (fp *fakeProcess) Kill() error {
	args := fp.Called()
	return args.Error(0)
}

func TestGroupMethodsWhenNil(t *testing.T) {
	var p *processGroup
	assert.Nil(t, p)
	assert.Equal(t, ErrNotConfigured, p.Run())
	assert.Equal(t, ErrNotConfigured, p.Signal(syscall.SIGHUP))
	proc, err := p.NewProcess("dir", "ls -al")
	assert.Nil(t, proc)
	assert.Equal(t, ErrNotConfigured, err)
}

func TestGroup(t *testing.T) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, err := New(ctx, WithStdOut(os.Stdout), WithStdErr(os.Stdout))
	assert.Nil(t, err)

	assert.Nil(t, WithEnv([]string{"FOO=BAR"})(g))
	assert.NotNil(t, WithEnv([]string{"FOO=BAR"})(g))

	assert.NotNil(t, WithStdOut(os.Stdout)(g))
	assert.NotNil(t, WithStdErr(os.Stdout)(g))

	port := ":6773"
	proc, err := g.NewProcess("server", "ls -al")
	assert.Nil(t, err)
	err = proc.AddEnv("PORT", port)
	assert.Nil(t, err)

	_, err = g.NewProcess("worker", "ls -al")
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

	_, err = g.NewProcess("env1", "env")
	assert.Nil(t, err)

	err = g.Run()
	assert.Nil(t, err)
	s := stdout.String()
	if !strings.Contains(s, envKey) {
		t.Fatalf("Env key %s is missing from the commands environment.", envKey)
	}
}

func TestRun(t *testing.T) {
	fake1 := new(fakeProcess)
	fake1.On("Run").Return(errors.New("foo"))
	fake2 := new(fakeProcess)
	fake2.On("Run").Return(nil)

	p := processGroup{
		processes: []Process{fake1, fake2},
	}

	err := p.Run()
	assert.NotNil(t, err)
	fake1.AssertCalled(t, "Run")
	fake2.AssertCalled(t, "Run")
}

func TestSignal(t *testing.T) {
	fake1 := new(fakeProcess)
	fake1.On("Run").Return(nil)
	fake1.On("Signal", syscall.SIGHUP).Return(errors.New("foo"))
	fake2 := new(fakeProcess)
	fake2.On("Run").Return(nil)
	fake2.On("Signal", syscall.SIGHUP).Return(nil)

	p := processGroup{
		processes: []Process{fake1, fake2},
	}

	go p.Run()
	err := p.Signal(syscall.SIGHUP)
	assert.NotNil(t, err)
	fake1.AssertCalled(t, "Signal", syscall.SIGHUP)
	fake2.AssertCalled(t, "Signal", syscall.SIGHUP)

}
