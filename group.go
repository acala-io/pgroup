package pgroup

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"sync"
	"syscall"

	"github.com/kvz/logstreamer"
	"golang.org/x/sync/errgroup"
)

var (
	// ErrNotConfigured is returned when process (or group) has not been properly configured
	ErrNotConfigured = errors.New("Process or Group not configured")
)

type groupOption func(*processGroup) error

type processGroup struct {
	configuration
	ctx       context.Context
	m         sync.Mutex
	processes []Process
}

func (p *processGroup) baseContext() context.Context {
	if p == nil || p.ctx == nil {
		return context.Background()
	}
	return p.ctx
}

func (p *processGroup) getOptions(name string) []processOption {

	options := []processOption{}
	if p == nil {
		return options
	}

	logger := log.New(p.stdout(), name+": ", log.Ldate|log.Ltime)

	logStreamerOut := logstreamer.NewLogstreamer(logger, "stdout", false)
	options = append(options, withStdOut(logStreamerOut))

	logStreamerErr := logstreamer.NewLogstreamer(logger, "stderr", true)
	options = append(options, withStdErr(logStreamerErr))

	if len(p.env) > 0 {
		options = append(options, withEnv(p.env))
	}

	return options
}

func (p *processGroup) checkConfigured() error {
	if p == nil || len(p.processes) == 0 {
		return ErrNotConfigured
	}
	return nil
}

func (p *processGroup) NewProcess(name, cmd string) error {

	if p == nil {
		return ErrNotConfigured
	}

	proc, err := newProcess(p.ctx, cmd, p.getOptions(name)...)
	if err != nil {
		return err
	}
	p.m.Lock()
	p.processes = append(p.processes, proc)
	p.m.Unlock()
	return nil
}

func (p *processGroup) Run() error {

	err := p.checkConfigured()
	if err != nil {
		return err
	}

	var g errgroup.Group
	for _, i := range p.processes {
		proc := i
		g.Go(func() error {
			return proc.Run()
		})
	}
	return g.Wait()
}

func (p *processGroup) Signal(s syscall.Signal) error {

	err := p.checkConfigured()
	if err != nil {
		return err
	}

	var g errgroup.Group
	for _, i := range p.processes {
		proc := i
		g.Go(func() error {
			return proc.Signal(s)
		})
	}
	return g.Wait()
}

func New(ctx context.Context, options ...groupOption) (*processGroup, error) {
	if ctx == nil {
		panic("nil Context")
	}

	var err error
	p := processGroup{
		ctx: ctx,
	}

	for _, o := range options {
		err = o(&p)
		if err != nil {
			return nil, err
		}
	}

	return &p, nil
}

func WithStdOut(w io.Writer) groupOption {
	return func(p *processGroup) error {
		if p.outWriter != nil {
			return errors.New("outWriter already configured")
		}
		p.outWriter = w
		return nil
	}
}

func WithStdErr(w io.Writer) groupOption {
	return func(p *processGroup) error {
		if p.errWriter != nil {
			return errors.New("errWriter already configured")
		}
		p.errWriter = w
		return nil
	}
}

func WithEnv(env []string) groupOption {
	return func(p *processGroup) error {
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

var _ ProcessGroup = (*processGroup)(nil)
