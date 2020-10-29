package scrimplb

import (
	"log"
	"math/rand"
	"time"
)

// PushTask runs a Pusher on a regular, config-defined basis
type PushTask struct {
	config       *ScrimpConfig
	sleepTime    time.Duration
	maxJitter    time.Duration
	failureCount int
}

// NewPushTask creates a new PushTask with the given config
func NewPushTask(config *ScrimpConfig) *PushTask {
	return &PushTask{
		config,
		config.LoadBalancerConfig.PushPeriod,
		config.LoadBalancerConfig.PushJitter,
		0,
	}
}

// Loop should be called in/as a goroutine and will regularly push state
func (p *PushTask) Loop() {
	for {
		if p.failureCount > 0 {
			backoffSleep := time.Second * 5 * time.Duration(p.failureCount)
			log.Printf("sleeping for %v extra due to previous failure\n", backoffSleep)
			time.Sleep(backoffSleep)
		}

		time.Sleep(p.sleepTime)

		randSleep := time.Duration(rand.Int63n(p.maxJitter.Nanoseconds())).Round(time.Millisecond)
		time.Sleep(randSleep)

		err := p.config.Provider.PushSeed(p.config.Resolver, p.config.PortRaw)

		if err != nil {
			log.Printf("failed to push seed: %v\n", err)
			p.failureCount++
		} else {
			p.failureCount = 0
		}
	}
}
