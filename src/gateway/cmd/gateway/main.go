package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gazizov-ai/lab2-rsoi/src/gateway/internal/circuitbreaker"
	"github.com/gazizov-ai/lab2-rsoi/src/gateway/internal/clients"
	"github.com/gazizov-ai/lab2-rsoi/src/gateway/internal/config"
	"github.com/gazizov-ai/lab2-rsoi/src/gateway/internal/httpserver"
	"github.com/gazizov-ai/lab2-rsoi/src/gateway/internal/service"
)

func main() {
	cfg := config.Load()

	loyaltyBreaker := circuitbreaker.New(10, 0.5, 5*time.Second)
	reservationBreaker := circuitbreaker.New(10, 0.5, 5*time.Second)
	paymentBreaker := circuitbreaker.New(10, 0.5, 5*time.Second)

	resClient := clients.NewReservationClient(cfg.ReservationURL, reservationBreaker)
	payClient := clients.NewPaymentClient(cfg.PaymentURL, paymentBreaker)
	loyalClient := clients.NewLoyaltyClient(cfg.LoyaltyURL, loyaltyBreaker)

	svc := service.NewGatewayService(resClient, payClient, loyalClient)
	router := httpserver.NewRouter(svc)

	log.Printf("gateway listening on %s", cfg.Addr())
	if err := http.ListenAndServe(cfg.Addr(), router); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
