package handlers

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	jwtSecret string
	pool      *pgxpool.Pool
}

func NewAuthHandler(secret string, pool *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{jwtSecret: secret, pool: pool}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	userID := req.Email
	if userID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "email required"})
	}

	// Auto-create profile if not exists
	h.pool.Exec(c.Context(),
		`INSERT INTO profiles (id, email, balance, onboarding_completed)
		 VALUES ($1, $1, 0, true)
		 ON CONFLICT (id) DO NOTHING`, userID)

	adminEmail := os.Getenv("ADMIN_EMAIL")
	isAdmin := userID == adminEmail

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "token generation failed"})
	}

	return c.JSON(fiber.Map{
		"token":    tokenString,
		"user_id":  userID,
		"email":    req.Email,
		"is_admin": isAdmin,
	})
}
