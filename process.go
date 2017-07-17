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

	errChan := make(chan error)
	go func() {
		errChan <- p.inner.Wait()
	}()

	if p.ctx == nil {
		return <-errChan
	}

	for {
		select {
		case <-p.ctx.Done():
			return p.Kill()
		case err := <-errChan:
			return err
		}
	}
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

	// TODO: find shell command splitter package
	sm := strings.Split(cmd, " ")

	p := &process{
		inner: exec.Command(sm[0], sm[1:]...),
		ctx:   ctx,
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

	sysProcAttr := new(syscall.SysProcAttr)
	sysProcAttr.Setpgid = true
	sysProcAttr.Pgid = 0

	p.inner.SysProcAttr = sysProcAttr

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
