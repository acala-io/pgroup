package pgroup

import "io"

// passThroughWriter implements an io.Writer but doesn't do anything.
type passThroughWriter struct {
}

func (d passThroughWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// Common utility configuration shared by process & process group
type configuration struct {
	env       []string
	outWriter io.Writer
	errWriter io.Writer
}

// wrapper around the stdout writer
func (p *configuration) stdout() io.Writer {
	if p == nil || p.outWriter != nil {
		return p.outWriter
	}
	return passThroughWriter{}
}

// wrapper around the stderr writer
func (p *configuration) stderr() io.Writer {
	if p == nil || p.errWriter != nil {
		return p.errWriter
	}
	return passThroughWriter{}
}
