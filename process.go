package pgroup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type processOption func(*process) error

type process struct {
	configuration
	ctx   context.Context
	inner *exec.Cmd
}

var _ Process = (*process)(nil)

func (p *process) Run() error {
	if p == nil {
		return ErrNotConfigured
	}
	if err := p.inner.Start(); err != nil {
		return err
	}
	return p.inner.Wait()
}

func (p *process) AddEnv(key, val string) error {
	if p == nil {
		return ErrNotConfigured
	}
	p.env = append(p.env, fmt.Sprintf("%s=%s", key, val))
	return nil
}

func (p *process) Signal(s syscall.Signal) error {
	if p == nil {
		return ErrNotConfigured
	}
	return syscall.Kill(p.inner.Process.Pid, s)
}

func (p *process) Kill() error {
	g := 0 - p.inner.Process.Pid
	return syscall.Kill(g, syscall.SIGKILL)
}

// newProcess creates & configures a new process
func newProcess(ctx context.Context, cmd string, options ...processOption) (*process, error) {

	var err error

	if ctx == nil {
		ctx = context.Background()
	}

	// TODO: find shell command splitter package
	sm := strings.Split(cmd, " ")

	p := &process{
		inner: exec.CommandContext(ctx, sm[0], sm[1:]...),
	}

	for _, o := range options {
		err = o(p)
		if err != nil {
			return nil, err
		}
	}

	if len(p.env) > 0 {
		p.inner.Env = p.env
	}

	p.inner.Stdout = p.stdout()
	p.inner.Stderr = p.stderr()

	return p, nil
}

// withStdOut returns a processOption setting the configuration structs outWriter
func withStdOut(w io.Writer) processOption {
	return func(p *process) error {
		if p.outWriter != nil {
			return errors.New("outWriter already configured")
		}
		p.outWriter = w
		return nil
	}
}

// withStdErr returns a processOption setting the configuration structs errWriter
func withStdErr(w io.Writer) processOption {
	return func(p *process) error {
		if p.errWriter != nil {
			return errors.New("errWriter already configured")
		}
		p.errWriter = w
		return nil
	}
}

// withEnv returns a processOption setting the configuration structs environment
func withEnv(env []string) processOption {
	return func(p *process) error {
		if len(p.env) > 0 {
			return errors.New("environment already configured")
		}

		e := os.Environ()

		for _, i := range env {
			e = append(e, i)
		}
		p.env = env
		return nil
	}
}
