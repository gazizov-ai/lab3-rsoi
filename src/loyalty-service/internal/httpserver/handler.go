package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gazizov-ai/lab2-rsoi/src/loyalty-service/internal/service"
)

type Handler struct {
	loyaltyService *service.LoyaltyService
}

func NewHandler(s *service.LoyaltyService) *Handler {
	return &Handler{loyaltyService: s}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := h.loyaltyService.Health(r.Context()); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) Loyalty(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 3 || parts[0] != "internal" || parts[1] != "loyalty" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isDecrement := len(parts) == 4 && parts[3] == "decrement"

	var username string
	if isDecrement {
		username = parts[2]
	} else if len(parts) == 3 {
		username = parts[2]
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if isDecrement {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		resp, err := h.loyaltyService.GetLoyalty(r.Context(), username)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "internal error"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)

	case http.MethodPost:
		var err error
		if isDecrement {
			err = h.loyaltyService.DecrementReservationCount(r.Context(), username)
		} else {
			err = h.loyaltyService.IncrementReservationCount(r.Context(), username)
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) Increment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	username := parts[len(parts)-2]
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.loyaltyService.IncrementReservationCount(r.Context(), username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
