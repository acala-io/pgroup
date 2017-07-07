package pgroup

import "syscall"

// Process encapsulates the basic methods a process requires
type Process interface {
	Run() error
	AddEnv(string, string) error
	Signal(s syscall.Signal) error
}

// ProcessGroup adds wrapper for a process' constructor
type ProcessGroup interface {
	Run() error
	Signal(s syscall.Signal) error
	NewProcess(name, cmd string) (Process, error)
}
