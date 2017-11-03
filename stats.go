package main

import "time"

type Stats struct {
	Start      time.Time `json:"start"`
	Stop       time.Time `json:"stop,omitempty"`
	Size       int64     `json:"size"`
	ExitCode   int       `json:"exit_code"`
	Stacktrace string    `json:"stacktrace"`
}

// NewStats creates Stats object
func NewStats() *Stats {
	return &Stats{
		Size:  0,
		Start: time.Now().UTC(),
	}
}
