package circuitbreaker

import (
	"sync"
	"time"
)

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

type CircuitBreaker struct {
	mu sync.Mutex

	state State

	failures  int
	successes int

	window []bool
	index  int

	windowSize       int
	failureThreshold float64
	openTimeout      time.Duration
	lastStateChange  time.Time
}

func New(windowSize int, threshold float64, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            Closed,
		window:           make([]bool, windowSize),
		windowSize:       windowSize,
		failureThreshold: threshold,
		openTimeout:      timeout,
		lastStateChange:  time.Now(),
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == Open {
		if time.Since(cb.lastStateChange) > cb.openTimeout {
			cb.state = HalfOpen
			cb.failures = 0
			cb.successes = 0
		} else {
			return false
		}
	}

	return true
}

func (cb *CircuitBreaker) Record(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	old := cb.window[cb.index]
	if old {
		cb.failures--
	} else {
		cb.successes--
	}

	cb.window[cb.index] = !success

	if success {
		cb.successes++
	} else {
		cb.failures++
	}

	cb.index = (cb.index + 1) % cb.windowSize

	cb.evaluate()
}

func (cb *CircuitBreaker) evaluate() {
	failureRate := float64(cb.failures) / float64(cb.windowSize)

	switch cb.state {
	case Closed:
		if failureRate >= cb.failureThreshold {
			cb.state = Open
			cb.lastStateChange = time.Now()
		}
	case HalfOpen:
		if cb.failures > 0 {
			cb.state = Open
			cb.lastStateChange = time.Now()
		} else if cb.successes > cb.windowSize/2 {
			cb.state = Closed
		}
	}
}
