# pgroup

[![Build Status](https://travis-ci.org/acala-io/pgroup.png?branch=master)](https://travis-ci.org/acala-io/pgroup)
[![Go Report Card](https://goreportcard.com/badge/github.com/acala-io/pgroup)](https://goreportcard.com/report/github.com/acala-io/pgroup)

Simple library for managing a group of processes.

## Example

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "github.com/acala-io/pgroup"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    // create process group with prefixed stdout & stderr writers
    grp, err := pgroup.New(ctx, WithStdOut(os.Stdout), WithStdErr(os.Stdout))
    if err != nil {
        log.Fatal(err)
    }

    _, err := grp.NewProcess("server", "python myserver.py")
    if err != nil {
        log.Fatal(err)
    }

    _, err := grp.NewProcess("worker", "python myworker.py")
    if err != nil {
        log.Fatal(err)
    }

    go grp.Run()

    grp.Signal(<-sigs)
}
```
