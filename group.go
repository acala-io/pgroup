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

// getOptions returns an array of processOption used for configurating processes.
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

// checkConfigured is called from functions which require a process to have
// been added. It returns an error if that is not the case.
func (p *processGroup) checkConfigured() error {
	if p == nil || len(p.processes) == 0 {
		return ErrNotConfigured
	}
	return nil
}

// NewProcess adds a new process with a name and a command
func (p *processGroup) NewProcess(name, cmd string, options ...processOption) (Process, error) {

	if p == nil {
		return nil, ErrNotConfigured
	}

	proc, err := newProcess(p.ctx, cmd, p.getOptions(name)...)
	if err != nil {
		return nil, err
	}
	p.m.Lock()
	// TODO: make processes a map instead because there is a key!
	p.processes = append(p.processes, proc)
	p.m.Unlock()
	return proc, nil
}

// Run runs a group of processes
func (p *processGroup) Run() error {

	// TODO: will dangling processes be running if the first one fails? write test!
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

// Signal propagates signal down to all the groups processes
func (p *processGroup) Signal(s syscall.Signal) error {

	// TODO: is it possible to guarantee that all processes get the signal, even if the first one errors? write test!
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

// New creates a new group of processes.
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

// WithStdOut configures an io.Writer for the groups combined stdout output.
func WithStdOut(w io.Writer) groupOption {
	return func(p *processGroup) error {
		if p.outWriter != nil {
			return errors.New("outWriter already configured")
		}
		p.outWriter = w
		return nil
	}
}

// WithStdErr configures an io.Writer for the groups combined stderr output.
func WithStdErr(w io.Writer) groupOption {
	return func(p *processGroup) error {
		if p.errWriter != nil {
			return errors.New("errWriter already configured")
		}
		p.errWriter = w
		return nil
	}
}

// WithEnv extends the groups configured environment variables.
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
