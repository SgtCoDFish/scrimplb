package pusher

import (
	"fmt"
	"math/rand"
	"time"
)

// Pusher is an interface which abstracts pushing IP data to a remote source to
// aid new instances joining the cluster
type Pusher interface {
	PushState() error
}

// PushTask runs a Pusher on a regular, config-defined basis
type PushTask struct {
	pusher                Pusher
	sleepTime             time.Duration
	maxJitterMilliseconds int64
}

// NewPushTask creates a new PushTask with the given config
func NewPushTask(pusher Pusher, sleepTime time.Duration, maxJitterMilliseconds int64) *PushTask {
	return &PushTask{
		pusher,
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

		err := p.pusher.PushState()

		if err != nil {
			fmt.Printf("failed to push state: %v\n", err)
		}
	}
}
