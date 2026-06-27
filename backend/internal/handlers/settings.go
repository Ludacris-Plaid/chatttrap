package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingsHandler struct {
	pool *pgxpool.Pool
}

func NewSettingsHandler(pool *pgxpool.Pool) *SettingsHandler {
	return &SettingsHandler{pool: pool}
}

func (h *SettingsHandler) GetSettings(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var countryCode, webhookURL string
	var ringtone, vibrate bool

	err := h.pool.QueryRow(c.Context(),
		`SELECT country_code, webhook_url, ringtone, vibrate FROM user_settings WHERE user_id=$1`, uid).
		Scan(&countryCode, &webhookURL, &ringtone, &vibrate)

	if err != nil {
		countryCode = "+1"
		webhookURL = ""
		ringtone = true
		vibrate = true
	}

	return c.JSON(fiber.Map{
		"country_code": countryCode,
		"webhook_url":  webhookURL,
		"ringtone":     ringtone,
		"vibrate":      vibrate,
	})
}

func (h *SettingsHandler) UpdateSettings(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req struct {
		CountryCode string `json:"country_code"`
		WebhookURL  string `json:"webhook_url"`
		Ringtone    bool   `json:"ringtone"`
		Vibrate     bool   `json:"vibrate"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid"})
	}

	_, err := h.pool.Exec(c.Context(), `
		INSERT INTO user_settings (user_id, country_code, webhook_url, ringtone, vibrate, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			country_code = EXCLUDED.country_code,
			webhook_url = EXCLUDED.webhook_url,
			ringtone = EXCLUDED.ringtone,
			vibrate = EXCLUDED.vibrate,
			updated_at = NOW()
	`, uid, req.CountryCode, req.WebhookURL, req.Ringtone, req.Vibrate)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "save failed"})
	}

	return c.JSON(fiber.Map{"status": "saved"})
}
