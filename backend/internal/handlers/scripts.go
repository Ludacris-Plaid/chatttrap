package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"hushcircuits/api/internal/database"
	"hushcircuits/api/internal/models"
	"hushcircuits/api/internal/services"
)

type ScriptHandler struct {
	q              *database.Queries
	pool           *pgxpool.Pool
	fsSvc          *services.FeatherlessService
	scriptCostTokens int
}

func NewScriptHandler(pool *pgxpool.Pool, fsSvc *services.FeatherlessService, scriptCostTokens int) *ScriptHandler {
	return &ScriptHandler{
		q:              database.NewQueries(pool),
		pool:           pool,
		fsSvc:          fsSvc,
		scriptCostTokens: scriptCostTokens,
	}
}

func (h *ScriptHandler) Generate(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req models.GenerateScriptRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if req.TargetName == "" || req.Goal == "" {
		return c.Status(400).JSON(fiber.Map{"error": "target_name and goal required"})
	}

	script, err := h.fsSvc.GenerateScript(&req)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "generation failed"})
	}

	id := uuid.New().String()
	if err := h.q.InsertScript(c.Context(), id, uid, req.TargetName, req.TargetAge, req.TargetService, req.TargetDetails, req.Goal, req.ScriptType, script, h.scriptCostTokens); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "store failed"})
	}

	return c.JSON(fiber.Map{"id": id, "script": script, "tokens_cost": h.scriptCostTokens})
}

func (h *ScriptHandler) List(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	rows, err := h.q.GetScripts(c.Context(), uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query failed"})
	}
	defer rows.Close()

	var scripts []fiber.Map
	for rows.Next() {
		var id, title, tname, svc, goal, stype, content string
		var tok int
		var lib bool
		var ca interface{}
		if err := rows.Scan(&id, &title, &tname, &svc, &goal, &stype, &content, &tok, &lib, &ca); err != nil {
			continue
		}
		scripts = append(scripts, fiber.Map{
			"id": id, "title": title, "target_name": tname, "target_service": svc, "goal": goal,
			"script_type": stype, "content": content, "tokens_cost": tok, "is_library": lib, "created_at": ca,
		})
	}
	if scripts == nil {
		scripts = []fiber.Map{}
	}
	return c.JSON(fiber.Map{"scripts": scripts})
}

func (h *ScriptHandler) Delete(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	id := c.Params("id")
	h.pool.Exec(c.Context(), `DELETE FROM scripts WHERE id = $1 AND user_id = $2`, id, uid)
	return c.JSON(fiber.Map{"status": "deleted"})
}
