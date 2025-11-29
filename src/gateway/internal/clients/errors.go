package clients

import "errors"

var ErrCircuitOpen = errors.New("circuit breaker open")
