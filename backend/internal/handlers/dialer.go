package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"hushcircuits/api/internal/database"
	"hushcircuits/api/internal/models"
	"hushcircuits/api/internal/phone"
	"hushcircuits/api/internal/services"
)

type DialerHandler struct {
	q       *database.Queries
	fsSvc   *services.FreeSwitchService
	cnam    *services.CNAMService
	tokens  *services.TokenService
	pool    *pgxpool.Pool
}

func NewDialerHandler(pool *pgxpool.Pool, fsSvc *services.FreeSwitchService, cnam *services.CNAMService, tokens *services.TokenService) *DialerHandler {
	return &DialerHandler{
		q:      database.NewQueries(pool),
		pool:   pool,
		fsSvc:  fsSvc,
		cnam:   cnam,
		tokens: tokens,
	}
}

func (h *DialerHandler) LookupCNAM(c *fiber.Ctx) error {
	number := c.Query("number")
	if number == "" {
		return c.Status(400).JSON(fiber.Map{"error": "number required"})
	}
	name, err := h.cnam.Lookup(number)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "cnam lookup failed"})
	}
	return c.JSON(models.CNAMResponse{Number: number, Name: name})
}

func (h *DialerHandler) OriginateCall(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req models.OriginateCallRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if req.DestinationNumber == "" || req.SpoofedCallerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "destination and caller id required"})
	}
	if !phone.Validate(req.SpoofedCallerID) || !phone.Validate(req.DestinationNumber) {
		// Allow SIP usernames (non-phone-number IDs like "10428")
		if len(req.SpoofedCallerID) < 3 || len(req.DestinationNumber) < 3 {
			return c.Status(400).JSON(fiber.Map{"error": "numbers must be 1+10 digit format: 1XXXXXXXXXX, or a valid SIP username"})
		}
	}
	req.SpoofedCallerID = phone.Normalize(req.SpoofedCallerID)
	req.DestinationNumber = phone.Normalize(req.DestinationNumber)

	balance, isVIP, _, err := h.q.GetProfile(c.Context(), uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server error"})
	}

	minCost := services.TokenRate
	if !isVIP && balance < minCost {
		return c.Status(402).JSON(fiber.Map{"error": "insufficient balance", "balance": balance, "required": minCost})
	}

	cost, usd := h.tokens.CalculateCost(60)
	if !isVIP {
		if err := h.q.ExecTx(c.Context(), func(tx pgx.Tx) error {
			return h.q.DeductBalanceTx(tx, c.Context(), uid, usd)
		}); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "balance deduction failed"})
		}
	}

	sipCallID, err := h.fsSvc.OriginateCall(&req)
	if err != nil {
		// SIP call failed — refund and create a mock call for testing
		if !isVIP {
			h.q.ExecTx(c.Context(), func(tx pgx.Tx) error {
				return h.q.AddBalanceTx(tx, c.Context(), uid, usd)
			})
		}
	}

	dbCallID := uuid.New().String()
	if err := h.q.InsertCall(c.Context(), dbCallID, uid, req.SpoofedCallerID, req.SpoofedName, req.DestinationNumber, "initiated", cost, usd); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to record call"})
	}

	sipErr := ""
	if err != nil {
		sipErr = err.Error()
	}
	return c.JSON(fiber.Map{
		"call_id":        dbCallID,
		"sip_call_id":    sipCallID,
		"status":         "initiated",
		"tokens_deducted": cost,
		"cost_usd":        usd,
		"sip_error":       sipErr,
	})
}

func (h *DialerHandler) EndCall(c *fiber.Ctx) error {
	callID := c.Params("callId")
	h.fsSvc.HangupCall(callID)

	_, _, _, _, _, _, _, dur, tok, cost, _, _, err := h.q.GetCall(c.Context(), callID)
	if err != nil {
		dur, tok, cost = 0, 1, services.TokenRate
	}
	if dur == 0 {
		dur = 1
	}
	if tok == 0 {
		tok = 1
	}
	if cost == 0 {
		cost = float64(tok) * services.TokenRate
	}

	h.q.UpdateCallEnded(c.Context(), callID, dur, tok, cost)
	return c.JSON(fiber.Map{"status": "completed", "duration_seconds": dur, "tokens_cost": tok, "cost_usd": cost})
}

func (h *DialerHandler) GetCall(c *fiber.Ctx) error {
	callID := c.Params("callId")
	id, uid, scid, sname, dest, status, dtmf, dur, tok, cost, ca, ea, err := h.q.GetCall(c.Context(), callID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "call not found"})
	}
	return c.JSON(fiber.Map{
		"id": id, "user_id": uid, "spoofed_caller_id": scid, "spoofed_name": sname,
		"destination_number": dest, "status": status, "duration_seconds": dur,
		"tokens_cost": tok, "cost_usd": cost, "dtmf_captured": dtmf,
		"created_at": ca, "ended_at": ea,
	})
}

func (h *DialerHandler) SubmitDTMF(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	callID := c.Params("callId")
	var req struct {
		Digit       string `json:"digit"`
		TimestampMs int    `json:"timestamp_ms"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid"})
	}
	if err := h.q.InsertDTMF(c.Context(), uuid.New().String(), callID, uid, req.Digit, req.TimestampMs); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "store failed"})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *DialerHandler) GetDTMF(c *fiber.Ctx) error {
	callID := c.Params("callId")
	digits, err := h.q.GetDTMFByCall(c.Context(), callID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query failed"})
	}
	return c.JSON(fiber.Map{"call_id": callID, "digits": digits})
}

func (h *DialerHandler) MuteCall(c *fiber.Ctx) error {
	callID := c.Params("callId")
	var req struct{ Muted bool `json:"muted"` }
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid"})
	}
	h.fsSvc.MuteCall(callID, req.Muted)
	return c.JSON(fiber.Map{"status": "ok", "muted": req.Muted})
}
