package main

import (
	"log"
	"time"
)

type Runner struct {
	Interval time.Duration
	Strategy Strategy
}

func (r *Runner) Start() {
	ticker := time.NewTicker(r.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println()
			if err := r.Strategy.Execute(); err != nil {
				log.Printf("Error executing strategy: %v", err)
			}
		}
	}
}

func NewRunner(interval time.Duration, strategy Strategy) *Runner {
	return &Runner{
		Interval: interval,
		Strategy: strategy,
	}
}
