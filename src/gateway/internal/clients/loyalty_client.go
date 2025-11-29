package clients

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gazizov-ai/lab2-rsoi/src/gateway/internal/circuitbreaker"
	"github.com/gazizov-ai/lab2-rsoi/src/gateway/internal/model"
)

type LoyaltyClient struct {
	baseURL string
	client  *http.Client
	breaker *circuitbreaker.CircuitBreaker
}

func NewLoyaltyClient(baseURL string, breaker *circuitbreaker.CircuitBreaker) *LoyaltyClient {
	return &LoyaltyClient{
		baseURL: baseURL,
		client:  &http.Client{},
		breaker: breaker,
	}
}

func (c *LoyaltyClient) GetLoyalty(username string) (model.Loyalty, error) {
	url := fmt.Sprintf("%s/internal/loyalty/%s", c.baseURL, username)

	if !c.breaker.Allow() {
		return model.Loyalty{}, ErrCircuitOpen
	}

	resp, err := c.client.Get(url)
	if err != nil {
		c.breaker.Record(false)
		return model.Loyalty{}, fmt.Errorf("request loyalty: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		c.breaker.Record(false)
	} else {
		c.breaker.Record(true)
	}

	var lo model.Loyalty
	if err := json.NewDecoder(resp.Body).Decode(&lo); err != nil {
		return model.Loyalty{}, fmt.Errorf("decode loyalty: %w", err)
	}

	return lo, nil
}

func (c *LoyaltyClient) IncrementReservation(username string) error {
	url := fmt.Sprintf("%s/internal/loyalty/%s", c.baseURL, username)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("build loyalty increment request: %w", err)
	}

	if !c.breaker.Allow() {
		return ErrCircuitOpen
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.breaker.Record(false)
		return fmt.Errorf("request loyalty increment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		c.breaker.Record(false)
	} else {
		c.breaker.Record(true)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("loyalty increment status %d", resp.StatusCode)
	}

	return nil
}
