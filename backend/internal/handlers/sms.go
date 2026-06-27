package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"hushcircuits/api/internal/database"
	"hushcircuits/api/internal/phone"
	"hushcircuits/api/internal/services"
)

type SMSHandler struct {
	q      *database.Queries
	pool   *pgxpool.Pool
	svc    *services.GENSMSService
}

func NewSMSHandler(pool *pgxpool.Pool, svc *services.GENSMSService) *SMSHandler {
	return &SMSHandler{
		q:    database.NewQueries(pool),
		pool: pool,
		svc:  svc,
	}
}

func (h *SMSHandler) SendSingle(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req struct {
		PhoneNumber string `json:"phone_number"`
		Content     string `json:"content"`
		SenderID    string `json:"sender_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid"})
	}
	if req.PhoneNumber == "" || req.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "phone_number and content required"})
	}
	if req.SenderID == "" {
		req.SenderID = "Service"
	}

	msgID, err := h.svc.Send(req.PhoneNumber, req.Content, req.SenderID)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "send failed"})
	}

	h.q.InsertSMSLog(c.Context(), uuid.New().String(), uid, req.PhoneNumber, req.Content, req.SenderID, msgID)
	return c.JSON(fiber.Map{"status": "sent", "message_id": msgID})
}

func (h *SMSHandler) SendBulk(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req struct {
		Targets  string `json:"targets"`
		Content  string `json:"content"`
		SenderID string `json:"sender_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid"})
	}

	phones := strings.Split(req.Targets, "\n")
	var cleaned []string
	for _, p := range phones {
		if p = strings.TrimSpace(p); p != "" {
			if phone.IsComplete(p) {
				cleaned = append(cleaned, phone.Normalize(p))
			} else {
				cleaned = append(cleaned, p)
			}
		}
	}

	sent, err := h.svc.SendBulk(cleaned, req.Content, req.SenderID)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "bulk send failed"})
	}

	id := uuid.New().String()
	h.q.InsertCampaign(c.Context(), id, uid, req.SenderID, req.Content, len(cleaned), sent)
	return c.JSON(fiber.Map{"campaign_id": id, "targets": len(cleaned), "sent": sent, "status": "completed"})
}
