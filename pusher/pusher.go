package pusher

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sgtcodfish/scrimplb/seed"
)

// PushTask runs a Pusher on a regular, config-defined basis
type PushTask struct {
	provider              seed.Provider
	sleepTime             time.Duration
	maxJitterMilliseconds int64
}

// NewPushTask creates a new PushTask with the given config
func NewPushTask(provider seed.Provider, sleepTime time.Duration, maxJitterMilliseconds int64) *PushTask {
	return &PushTask{
		provider,
		sleepTime,
		maxJitterMilliseconds,
	}
}

// Loop should be called in/as a goroutine and will regularly push state
func (p *PushTask) Loop() {
	for {
		time.Sleep(time.Second * p.sleepTime)

		if p.maxJitterMilliseconds != 0 {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(p.maxJitterMilliseconds)))
		}

		err := p.provider.PushSeed()

		if err != nil {
			fmt.Printf("failed to push seed: %v\n", err)
		}
	}
}
