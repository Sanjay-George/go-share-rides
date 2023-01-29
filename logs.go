package main

import "fmt"

type Logger struct {
	isEnabled bool
}

func (log *Logger) Log(message string) {
	if !log.isEnabled {
		return
	}

	fmt.Print(message)
}
