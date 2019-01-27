package worker

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sgtcodfish/scrimplb/seed"
)

// PushTask runs a Pusher on a regular, config-defined basis
type PushTask struct {
	provider  seed.Provider
	sleepTime time.Duration
	maxJitter time.Duration
}

// NewPushTask creates a new PushTask with the given config
func NewPushTask(provider seed.Provider, sleepTime time.Duration, maxJitter time.Duration) *PushTask {
	return &PushTask{
		provider,
		sleepTime,
		maxJitter,
	}
}

// Loop should be called in/as a goroutine and will regularly push state
func (p *PushTask) Loop() {
	for {
		time.Sleep(time.Second * p.sleepTime)

		randMs := time.Duration(rand.Int63n(p.maxJitter.Nanoseconds())).Round(time.Millisecond)
		time.Sleep(time.Nanosecond * randMs)

		err := p.provider.PushSeed()

		if err != nil {
			fmt.Printf("failed to push seed: %v\n", err)
		}
	}
}
