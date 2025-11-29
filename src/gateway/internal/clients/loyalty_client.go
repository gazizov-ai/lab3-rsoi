package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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
	endpoint := fmt.Sprintf("%s/internal/loyalty/%s", c.baseURL, url.PathEscape(username))

	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build loyalty increment request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request loyalty increment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("loyalty increment status %d", resp.StatusCode)
	}

	return nil
}

func (c *LoyaltyClient) DecrementReservation(username string) error {
	endpoint := fmt.Sprintf("%s/internal/loyalty/%s/decrement", c.baseURL, url.PathEscape(username))
	fmt.Println("CALL DEC LOYALTY:", endpoint)

	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build loyalty decrement request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request loyalty decrement: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("loyalty decrement status %d", resp.StatusCode)
	}

	return nil
}
