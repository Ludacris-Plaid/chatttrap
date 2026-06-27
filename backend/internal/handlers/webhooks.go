package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"hushcircuits/api/internal/database"
	"hushcircuits/api/internal/models"
	"hushcircuits/api/internal/services"
)

type WebhookHandler struct {
	q      *database.Queries
	pool   *pgxpool.Pool
	pmtSvc *services.NowPaymentsService
}

func NewWebhookHandler(pool *pgxpool.Pool, pmtSvc *services.NowPaymentsService) *WebhookHandler {
	return &WebhookHandler{
		q:      database.NewQueries(pool),
		pool:   pool,
		pmtSvc: pmtSvc,
	}
}

func (h *WebhookHandler) NowPayments(c *fiber.Ctx) error {
	sig := c.Get("x-nowpayments-sig")
	body := c.Body()

	if !h.pmtSvc.VerifyWebhook(body, sig) {
		return c.Status(401).JSON(fiber.Map{"error": "invalid signature"})
	}

	var wh models.NowPaymentsWebhook
	if err := c.BodyParser(&wh); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	userID := wh.OrderID
	tokens := int(wh.PriceAmount / services.TokenRate)

	// Update payment record
	h.pool.Exec(c.Context(),
		`UPDATE payments SET status = $1, txid = $2, confirmed_at = NOW() WHERE payment_id = $3`,
		wh.PaymentStatus, wh.TxID, wh.PaymentID)

	if wh.PaymentStatus == "finished" {
		h.q.AddBalance(c.Context(), userID, wh.PriceAmount)
		h.q.InsertTransaction(c.Context(), uuid.New().String(), userID, "deposit", wh.PriceAmount, tokens, "NowPayments: "+wh.PaymentID, "completed")

		// First deposit bonus
		var firstDeposit bool
		h.pool.QueryRow(c.Context(), `SELECT first_deposit FROM profiles WHERE id = $1`, userID).Scan(&firstDeposit)
		if !firstDeposit {
			bonus := wh.PriceAmount * 0.10
			h.q.AddBalance(c.Context(), userID, bonus)
			h.q.InsertTransaction(c.Context(), uuid.New().String(), userID, "deposit_bonus", bonus, int(bonus/0.50), "First deposit bonus (10%)", "completed")
			h.pool.Exec(c.Context(), `UPDATE profiles SET first_deposit = true WHERE id = $1`, userID)
		}
	}

	return c.JSON(fiber.Map{"status": "ok"})
}
