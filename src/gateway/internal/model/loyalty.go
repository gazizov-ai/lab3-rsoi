package model

type Loyalty struct {
	Status           string `json:"status,omitempty"`
	Discount         int    `json:"discount,omitempty"`
	ReservationCount int    `json:"reservationCount,omitempty"`
}
