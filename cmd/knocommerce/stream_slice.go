package main

import "time"

type StreamSlice struct {
	StartTime  time.Time
	EndTime    time.Time
	DateFormat string
	Duration   time.Duration
}
