package routes

import "github.com/gofiber/fiber/v2"

func Home(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Users API functioning properly."})
}
