package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gazizov-ai/lab2-rsoi/src/gateway/internal/circuitbreaker"
	"github.com/gazizov-ai/lab2-rsoi/src/gateway/internal/model"
)

type ReservationClient struct {
	baseURL string
	client  *http.Client
	breaker *circuitbreaker.CircuitBreaker
}

func NewReservationClient(baseURL string, breaker *circuitbreaker.CircuitBreaker) *ReservationClient {
	return &ReservationClient{
		baseURL: baseURL,
		client:  &http.Client{},
		breaker: breaker,
	}
}

func (c *ReservationClient) ListHotels(page, size int) (model.HotelsPage, error) {
	u, err := url.Parse(c.baseURL + "/internal/hotels")
	if err != nil {
		return model.HotelsPage{}, err
	}

	q := u.Query()
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if size > 0 {
		q.Set("size", strconv.Itoa(size))
	}
	u.RawQuery = q.Encode()

	if !c.breaker.Allow() {
		return model.HotelsPage{}, ErrCircuitOpen
	}

	resp, err := c.client.Get(u.String())
	if err != nil {
		c.breaker.Record(false)
		return model.HotelsPage{}, fmt.Errorf("list hotels: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		c.breaker.Record(false)
	} else {
		c.breaker.Record(true)
	}

	if resp.StatusCode != http.StatusOK {
		return model.HotelsPage{}, fmt.Errorf("list hotels status %d", resp.StatusCode)
	}

	var out model.HotelsPage
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return model.HotelsPage{}, fmt.Errorf("decode hotels: %w", err)
	}

	return out, nil
}

func (c *ReservationClient) GetHotel(hotelUID string) (model.Hotel, error) {
	url := fmt.Sprintf("%s/internal/hotels/%s", c.baseURL, hotelUID)

	if !c.breaker.Allow() {
		return model.Hotel{}, ErrCircuitOpen
	}

	resp, err := c.client.Get(url)
	if err != nil {
		c.breaker.Record(false)
		return model.Hotel{}, fmt.Errorf("get hotel: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.breaker.Record(true)
		return model.Hotel{}, nil
	}
	if resp.StatusCode >= 500 {
		c.breaker.Record(false)
		return model.Hotel{}, fmt.Errorf("get hotel status %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return model.Hotel{}, fmt.Errorf("get hotel status %d", resp.StatusCode)
	}

	c.breaker.Record(true)

	var h model.Hotel
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		return model.Hotel{}, fmt.Errorf("decode hotel: %w", err)
	}

	return h, nil
}

func (c *ReservationClient) CreateReservation(req model.ReservationInternal) (model.ReservationFull, error) {
	var body = struct {
		Username   string `json:"username"`
		HotelUID   string `json:"hotelUid"`
		StartDate  string `json:"startDate"`
		EndDate    string `json:"endDate"`
		PaymentUID string `json:"paymentUid"`
	}{
		Username:   req.Username,
		HotelUID:   req.HotelUID,
		StartDate:  req.StartDate.Format("2006-01-02"),
		EndDate:    req.EndDate.Format("2006-01-02"),
		PaymentUID: req.PaymentUID,
	}

	data, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/internal/reservations", c.baseURL)

	resp, err := c.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return model.ReservationFull{}, fmt.Errorf("create reservation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.ReservationFull{}, fmt.Errorf("reservation status %d", resp.StatusCode)
	}

	var out model.ReservationFull

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return model.ReservationFull{}, fmt.Errorf("decode reservation: %w", err)
	}

	return out, nil
}

func (c *ReservationClient) GetReservation(uid string) (model.ReservationFull, error) {
	url := fmt.Sprintf("%s/internal/reservations/%s", c.baseURL, uid)

	if !c.breaker.Allow() {
		return model.ReservationFull{}, ErrCircuitOpen
	}

	resp, err := c.client.Get(url)
	if err != nil {
		c.breaker.Record(false)
		return model.ReservationFull{}, fmt.Errorf("get reservation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.breaker.Record(true)
		return model.ReservationFull{}, nil
	}
	if resp.StatusCode >= http.StatusInternalServerError {
		c.breaker.Record(false)
		return model.ReservationFull{}, fmt.Errorf("reservation status %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return model.ReservationFull{}, fmt.Errorf("reservation status %d", resp.StatusCode)
	}

	c.breaker.Record(true)

	var out model.ReservationFull
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return model.ReservationFull{}, fmt.Errorf("decode reservation: %w", err)
	}

	return out, nil
}

func (c *ReservationClient) GetReservationsByUser(username string) ([]model.ReservationFull, error) {
	url := fmt.Sprintf("%s/internal/reservations/byUser/%s", c.baseURL, username)

	if !c.breaker.Allow() {
		return nil, ErrCircuitOpen
	}

	resp, err := c.client.Get(url)
	if err != nil {
		c.breaker.Record(false)
		return nil, fmt.Errorf("list reservations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		c.breaker.Record(false)
		return nil, fmt.Errorf("reservation status %d", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reservations status %d", resp.StatusCode)
	}

	c.breaker.Record(true)

	var out []model.ReservationFull
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode reservations: %w", err)
	}

	return out, nil
}

func (c *ReservationClient) CancelReservation(uid string) error {
	req, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/internal/reservations/%s", c.baseURL, uid),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("cancel reservation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cancel reservation status %d", resp.StatusCode)
	}

	return nil
}
