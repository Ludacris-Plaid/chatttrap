package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"hushcircuits/api/internal/database"
	"hushcircuits/api/internal/models"
	"hushcircuits/api/internal/services"
)

type WalletHandler struct {
	q           *database.Queries
	pool        *pgxpool.Pool
	pmtSvc      *services.NowPaymentsService
	vouchSvc    *services.VoucherService
	tokenRate   float64
	vipPrice    float64
	vipDays     int
}

func NewWalletHandler(pool *pgxpool.Pool, pmtSvc *services.NowPaymentsService, vouchSvc *services.VoucherService, tokenRate, vipPrice float64, vipDays int) *WalletHandler {
	return &WalletHandler{
		q:        database.NewQueries(pool),
		pool:     pool,
		pmtSvc:   pmtSvc,
		vouchSvc: vouchSvc,
		tokenRate: tokenRate,
		vipPrice:  vipPrice,
		vipDays:   vipDays,
	}
}

func (h *WalletHandler) GetBalance(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	balance, isVIP, vipExp, _ := h.q.GetProfile(c.Context(), uid)
	return c.JSON(fiber.Map{"balance": balance, "is_vip": isVIP, "vip_expires": vipExp})
}

func (h *WalletHandler) CreatePayment(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req struct {
		Currency string  `json:"currency"`
		Amount   float64 `json:"amount"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid"})
	}
	if req.Amount < 10 {
		return c.Status(400).JSON(fiber.Map{"error": "minimum $10"})
	}

	pmt, err := h.pmtSvc.CreatePayment(uid, &models.CreatePaymentRequest{
		Currency: req.Currency,
		Amount:   req.Amount,
	})
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "payment failed"})
	}
	return c.JSON(pmt)
}

func (h *WalletHandler) RedeemVoucher(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req struct{ Code string `json:"code"` }
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid"})
	}

	tokens, err := h.vouchSvc.Redeem(c.Context(), req.Code)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	amount := float64(tokens) * h.tokenRate
	h.q.AddBalance(c.Context(), uid, amount)
	h.q.InsertTransaction(c.Context(), uuid.New().String(), uid, "voucher", amount, tokens, "Voucher: "+req.Code, "completed")

	return c.JSON(fiber.Map{"status": "redeemed", "tokens": tokens})
}

func (h *WalletHandler) UpgradeVIP(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	balance, _, _, _ := h.q.GetProfile(c.Context(), uid)
	if balance < h.vipPrice {
		return c.Status(402).JSON(fiber.Map{"error": "insufficient", "required": h.vipPrice})
	}
	h.q.UpgradeVIP(c.Context(), uid, h.vipPrice, h.vipDays)
	return c.JSON(fiber.Map{"status": "vip_active", "duration_days": h.vipDays})
}

func (h *WalletHandler) GetTransactions(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	rows, err := h.q.GetTransactions(c.Context(), uid, 50)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query failed"})
	}
	defer rows.Close()

	var txns []fiber.Map
	for rows.Next() {
		var id, ttype, status, desc string
		var amount float64
		var tokens int
		var ca interface{}
		if err := rows.Scan(&id, &ttype, &amount, &tokens, &status, &desc, &ca); err != nil {
			continue
		}
		txns = append(txns, fiber.Map{"id": id, "type": ttype, "amount": amount, "tokens": tokens, "status": status, "description": desc, "created_at": ca})
	}
	if txns == nil {
		txns = []fiber.Map{}
	}
	return c.JSON(fiber.Map{"transactions": txns})
}
