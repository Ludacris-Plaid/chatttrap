package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var (
	jwksCache     interface{}
	jwksCacheTime time.Time
	jwksMu        sync.RWMutex
)

type jwksKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type jwksResponse struct {
	Keys []jwksKey `json:"keys"`
}

func Logger(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()
	slog.Info("request",
		"method", c.Method(),
		"path", c.Path(),
		"status", c.Response().StatusCode(),
		"duration", time.Since(start).String(),
	)
	return err
}

func Recovery(c *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic", "error", r)
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}
	}()
	return c.Next()
}

func CORS(origins []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")
		allowed := false
		for _, o := range origins {
			if o == "*" || origin == o {
				allowed = true
				if o == "*" {
					c.Set("Access-Control-Allow-Origin", "*")
				} else {
					c.Set("Access-Control-Allow-Origin", origin)
				}
				break
			}
		}
		if !allowed {
			if c.Method() == "OPTIONS" {
				return c.SendStatus(fiber.StatusNoContent)
			}
			return c.Next()
		}
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		c.Set("Access-Control-Allow-Credentials", "true")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.Next()
	}
}

func fetchJWKS() (*jwksResponse, error) {
	jwksMu.RLock()
	if jwksCache != nil && time.Since(jwksCacheTime) < 1*time.Hour {
		defer jwksMu.RUnlock()
		return jwksCache.(*jwksResponse), nil
	}
	jwksMu.RUnlock()

	jwksMu.Lock()
	defer jwksMu.Unlock()

	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		return nil, fmt.Errorf("SUPABASE_URL not set")
	}

	resp, err := http.Get(supabaseURL + "/auth/v1/.well-known/jwks.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	jwksCache = &jwks
	jwksCacheTime = time.Now()
	return &jwks, nil
}

func base64URLDecode(s string) ([]byte, error) {
	// JWT uses unpadded base64url
	return base64.RawURLEncoding.DecodeString(s)
}

func verifySupabaseToken(tokenString string) (*jwt.Token, error) {
	// Parse token without verification to get header
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode header
	headerBytes, err := base64URLDecode(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	slog.Debug("verifySupabaseToken", "alg", header.Alg, "kid", header.Kid)

	// Only handle ES256 (Supabase's current algorithm)
	if header.Alg != "ES256" {
		return nil, fmt.Errorf("unsupported algorithm: %s", header.Alg)
	}

	// Fetch JWKS
	jwks, err := fetchJWKS()
	if err != nil {
		return nil, err
	}

	slog.Debug("fetched JWKS", "key_count", len(jwks.Keys))

	// Find matching key
	var matchingKey *jwksKey
	for i := range jwks.Keys {
		if jwks.Keys[i].Kid == header.Kid {
			matchingKey = &jwks.Keys[i]
			break
		}
	}
	if matchingKey == nil {
		return nil, fmt.Errorf("no matching key found for kid: %s", header.Kid)
	}

	// Reconstruct ECDSA public key
	xBytes, err := base64URLDecode(matchingKey.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode X: %w", err)
	}
	yBytes, err := base64URLDecode(matchingKey.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Y: %w", err)
	}

	slog.Debug("decoded key coordinates", "x_len", len(xBytes), "y_len", len(yBytes))

	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)

	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	// Verify signature
	sigBytes, err := base64URLDecode(parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	slog.Debug("signature bytes", "length", len(sigBytes))

	// ECDSA signature is r || s, each 32 bytes for P-256
	if len(sigBytes) != 64 {
		return nil, fmt.Errorf("invalid signature length: %d", len(sigBytes))
	}
	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	// Hash the input (header.payload)
	message := parts[0] + "." + parts[1]
	hash := sha256.Sum256([]byte(message))

	slog.Debug("verifying ECDSA signature", "message_len", len(message), "hash_hex", fmt.Sprintf("%x", hash))

	if !ecdsa.Verify(pubKey, hash[:], r, s) {
		return nil, fmt.Errorf("signature verification failed")
	}

	// Parse claims
	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	var claims jwt.MapClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	// Check expiry
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}
	}

	slog.Debug("token verified successfully via JWKS")
	return &jwt.Token{Claims: claims, Valid: true}, nil
}

func Auth(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			auth = c.Get("X-API-Key")
		}

		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization"})
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")

		// 1. Try local HMAC JWT first
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		// 2. If local JWT fails, try Supabase JWKS verification
		if err != nil || !token.Valid {
			slog.Debug("local JWT verification failed, trying Supabase JWKS", "error", err)
			supabaseURL := os.Getenv("SUPABASE_URL")
			if supabaseURL != "" {
				token, err = verifySupabaseToken(tokenString)
			}
		}

		if err != nil || !token.Valid {
			slog.Warn("token verification failed", "error", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid claims"})
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user id not found in token"})
		}

		c.Locals("user_id", userID)
		c.Locals("claims", claims)
		return c.Next()
	}
}

func AdminEmail(adminEmail string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin only"})
		}
		// In demo mode, the user_id IS the email. In Supabase mode, check email claim.
		claims, _ := c.Locals("claims").(jwt.MapClaims)
		if claims != nil {
			if email, ok := claims["email"].(string); ok {
				if email == adminEmail {
					return c.Next()
				}
			}
		}
		if userID == adminEmail {
			return c.Next()
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin only"})
	}
}
