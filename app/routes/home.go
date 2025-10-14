package routes

import "github.com/gofiber/fiber/v2"

// Home godoc
// @Summary      Ping the Users API
// @Description  Confirms that the Users API is running and responding to requests.
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string "Ping response message"
// @Router       / [get]

func Home(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Users API functioning properly."})
}
