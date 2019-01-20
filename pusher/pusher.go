package pusher

import (
	"fmt"
	"math/rand"
	"time"
)

type Pusher struct {
	sleepTime             time.Duration
	maxJitterMilliseconds int64
}

func NewPusher(sleepTime time.Duration, maxJitterMilliseconds int64) *Pusher {
	return &Pusher{
		sleepTime,
		maxJitterMilliseconds,
	}
}

func (p *Pusher) Loop() {
	for {
		time.Sleep(time.Second * p.sleepTime)

		if p.maxJitterMilliseconds != 0 {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(p.maxJitterMilliseconds)))
		}

		doPush()
	}
}

func doPush() {
	fmt.Println("pushing")
}
