package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"hushcircuits/api/internal/database"
	"hushcircuits/api/internal/services"
)

type AdminHandler struct {
	q        *database.Queries
	pool     *pgxpool.Pool
	vouchSvc *services.VoucherService
}

func NewAdminHandler(pool *pgxpool.Pool, vouchSvc *services.VoucherService) *AdminHandler {
	return &AdminHandler{
		q:        database.NewQueries(pool),
		pool:     pool,
		vouchSvc: vouchSvc,
	}
}

func (h *AdminHandler) Dashboard(c *fiber.Ctx) error {
	var tu, ac, tc int
	var rev float64
	h.pool.QueryRow(c.Context(), `SELECT COUNT(*) FROM profiles`).Scan(&tu)
	h.pool.QueryRow(c.Context(), `SELECT COUNT(*) FROM calls WHERE status='initiated'`).Scan(&ac)
	h.pool.QueryRow(c.Context(), `SELECT COUNT(*) FROM calls`).Scan(&tc)
	h.pool.QueryRow(c.Context(), `SELECT COALESCE(SUM(cost_usd),0) FROM calls`).Scan(&rev)
	return c.JSON(fiber.Map{"total_users": tu, "active_calls": ac, "total_calls": tc, "total_revenue": rev})
}

func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	rows, err := h.pool.Query(c.Context(),
		`SELECT id, email, balance, is_vip, tokens_used, total_calls, created_at FROM profiles ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query failed"})
	}
	defer rows.Close()

	var users []fiber.Map
	for rows.Next() {
		var id, email string
		var bal float64
		var vip bool
		var tu, tc int64
		var ca interface{}
		if err := rows.Scan(&id, &email, &bal, &vip, &tu, &tc, &ca); err != nil {
			continue
		}
		users = append(users, fiber.Map{"id": id, "email": email, "balance": bal, "is_vip": vip, "tokens_used": tu, "total_calls": tc, "created_at": ca})
	}
	if users == nil {
		users = []fiber.Map{}
	}
	return c.JSON(fiber.Map{"users": users})
}

func (h *AdminHandler) AdjustBalance(c *fiber.Ctx) error {
	uid := c.Params("userId")
	var req struct {
		Amount float64 `json:"amount"`
		Reason string  `json:"reason"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid"})
	}

	if req.Amount >= 0 {
		h.q.AddBalance(c.Context(), uid, req.Amount)
	} else {
		h.q.DeductBalance(c.Context(), uid, -req.Amount)
	}
	h.q.InsertTransaction(c.Context(), uuid.New().String(), uid, "admin_adjustment", req.Amount, int(req.Amount/services.TokenRate), req.Reason, "completed")

	return c.JSON(fiber.Map{"status": "adjusted"})
}

func (h *AdminHandler) GenerateVoucher(c *fiber.Ctx) error {
	var req struct{ Prefix string `json:"prefix"` }
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid"})
	}
	if req.Prefix == "" {
		req.Prefix = "HUSH-25"
	}

	code, tokens, err := h.vouchSvc.Generate(req.Prefix)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "generation failed"})
	}

	h.pool.Exec(c.Context(), `INSERT INTO vouchers (id, code, tokens) VALUES ($1, $2, $3)`, uuid.New().String(), code, tokens)
	return c.JSON(fiber.Map{"code": code, "tokens": tokens})
}

func (h *AdminHandler) GetAllDTMFLogs(c *fiber.Ctx) error {
	rows, err := h.pool.Query(c.Context(),
		`SELECT dl.id, dl.call_id, dl.user_id, dl.digit, dl.timestamp_ms, dl.created_at,
		        c.spoofed_caller_id, c.destination_number
		 FROM dtmf_logs dl JOIN calls c ON c.id = dl.call_id
		 ORDER BY dl.created_at DESC LIMIT 200`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query failed"})
	}
	defer rows.Close()

	var logs []fiber.Map
	for rows.Next() {
		var id, cid, uid, dig, scid, dest string
		var ts int
		var ca interface{}
		rows.Scan(&id, &cid, &uid, &dig, &ts, &ca, &scid, &dest)
		logs = append(logs, fiber.Map{"id": id, "call_id": cid, "user_id": uid, "digit": dig, "timestamp_ms": ts, "created_at": ca, "spoofed_caller_id": scid, "destination": dest})
	}
	if logs == nil {
		logs = []fiber.Map{}
	}
	return c.JSON(fiber.Map{"dtmf_logs": logs})
}
