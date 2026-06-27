package handlers

import "github.com/gofiber/fiber/v2"

func userID(c *fiber.Ctx) (string, bool) {
	id, ok := c.Locals("user_id").(string)
	return id, ok
}
