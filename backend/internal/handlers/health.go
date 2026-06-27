package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool           *pgxpool.Pool
	gensmsKey      string
	featherlessKey string
	cnamKey        string
	nowpaymentsKey string
	sipRegistered  func() bool
}

func NewHealthHandler(pool *pgxpool.Pool, gensmsKey, featherlessKey, cnamKey, nowpaymentsKey string, sipRegistered func() bool) *HealthHandler {
	return &HealthHandler{
		pool:           pool,
		gensmsKey:      gensmsKey,
		featherlessKey: featherlessKey,
		cnamKey:        cnamKey,
		nowpaymentsKey: nowpaymentsKey,
		sipRegistered:  sipRegistered,
	}
}

type serviceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Mode   string `json:"mode"`
}

func (h *HealthHandler) GetServices(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	services := []serviceStatus{}

	// Database
	dbStatus := "connected"
	if err := h.pool.Ping(ctx); err != nil {
		dbStatus = "error"
	}
	services = append(services, serviceStatus{Name: "PostgreSQL", Status: dbStatus, Mode: "local"})

	// GenSMS
	gensmsMode := "mock"
	if h.gensmsKey != "" && h.gensmsKey != "demo" {
		gensmsMode = "live"
	}
	services = append(services, serviceStatus{Name: "GenSMS", Status: "ready", Mode: gensmsMode})

	// Featherless
	featherMode := "mock"
	if h.featherlessKey != "" && h.featherlessKey != "demo" {
		featherMode = "live"
	}
	services = append(services, serviceStatus{Name: "Featherless AI", Status: "ready", Mode: featherMode})

	// CNAM
	cnamMode := "mock"
	if h.cnamKey != "" && h.cnamKey != "demo" {
		cnamMode = "live"
	}
	services = append(services, serviceStatus{Name: "CNAM Lookup", Status: "ready", Mode: cnamMode})

	// NowPayments
	payMode := "mock"
	if h.nowpaymentsKey != "" {
		payMode = "live"
	}
	services = append(services, serviceStatus{Name: "NowPayments", Status: "ready", Mode: payMode})

	// SIP
	sipMode := "mock"
	sipStatus := "disconnected"
	if h.sipRegistered != nil {
		sipMode = "live"
		if h.sipRegistered() {
			sipStatus = "registered"
		} else {
			sipStatus = "unregistered"
		}
	}
	services = append(services, serviceStatus{Name: "SIP Trunk", Status: sipStatus, Mode: sipMode})

	return c.JSON(fiber.Map{"services": services})
}
