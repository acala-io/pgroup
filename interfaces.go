package pgroup

import "syscall"

// Process encapsulates the basic methods a process requires
type Process interface {
	Run() error
	Signal(s syscall.Signal) error
}

// ProcessGroup adds wrapper for a process' constructor
type ProcessGroup interface {
	Process
	NewProcess(name, cmd string) error
}
