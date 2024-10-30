clock
=====

[![Go Report Card](https://goreportcard.com/badge/github.com/bangzek/clock)](https://goreportcard.com/report/github.com/bangzek/clock)
[![Go Reference](https://pkg.go.dev/badge/github.com/bangzek/clock.svg)](https://pkg.go.dev/github.com/bangzek/clock)

Clock is a small library for mocking time in Go. It provides an interface
around the standard library's [`time`](https://pkg.go.dev/time) package so that
the application can use the realtime clock while tests can use the mock clock.
