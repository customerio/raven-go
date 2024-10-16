# raven 

## This was forked from github.com/getsentry/raven-go

These links do not apply to this fork:

[![Build Status](https://api.travis-ci.org/getsentry/raven-go.svg?branch=master)](https://travis-ci.org/getsentry/raven-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/getsentry/raven-go)](https://goreportcard.com/report/github.com/getsentry/raven-go)
[![GoDoc](https://godoc.org/github.com/getsentry/raven-go?status.svg)](https://godoc.org/github.com/getsentry/raven-go)

---

> The `raven-go` SDK is no longer maintained and was superseded by the `sentry-go` SDK.
> Learn more about the project on [GitHub](https://github.com/getsentry/sentry-go) and check out the [migration guide](https://docs.sentry.io/platforms/go/migration/).

---

raven is the official Go SDK for the [Sentry](https://github.com/getsentry/sentry)
event/error logging system.

- [**API Documentation**](https://godoc.org/github.com/getsentry/raven-go)
- [**Usage and Examples**](https://docs.sentry.io/clients/go/)

## Installation

```text
go get github.com/customerio/raven-go
```

Note: Go 1.22 and newer are supported. Earlier and newer versions may work, but have not been tested with this fork.

## Testing

```bash
go test .
```

Unfortunately this results in a few test failures. Since this fork has minimally changed, I do not consider these blockers for shipping the module. But they do need fixing. They are probably related to the switch to a modern go module.

```
--- FAIL: TestFunctionName (0.00s)
    stacktrace_test.go:50: incorrect package; got github.com/customerio/raven-go, want .
--- FAIL: TestStacktraceFrame (0.00s)
    stacktrace_test.go:84: incorrect Module: github.com/customerio/raven-go
    stacktrace_test.go:87: incorrect Lineno: 18
    stacktrace_test.go:90: expected InApp to be true
--- FAIL: TestStacktraceErrorsWithStack (0.00s)
    stacktrace_test.go:141: incorrect Module: github.com/customerio/raven-go
    stacktrace_test.go:150: incorrect Module: github.com/customerio/raven-go
```
