package routes

import (
	"errors"
	"users-api/app/appdata"
	"users-api/app/models"
	"users-api/app/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func MarkChapterAsRead(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)

	// Deserialize JSON body
	var req struct {
		Book    uint `json:"book"`
		Chapter uint `json:"chapter"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Book < 1 || req.Book > 66 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid book",
		})
	}

	bookStruct := appdata.Books[req.Book]

	if req.Chapter > bookStruct.Chapters {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chapter number",
		})
	}
	readHistory := models.ReadHistory{UserID: userID, Book: req.Book, Chapter: req.Chapter}
	result := appdata.DB.Create(&readHistory)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			// Successful response is intentional
			return c.JSON(fiber.Map{
				"error": "This chapter is already marked read",
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Something went wrong, try again later",
			})
		}
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Chapter marked read",
	})
}

func MarkChapterAsUnread(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	// Deserialize JSON body
	var req struct {
		Book    uint `json:"book"`
		Chapter uint `json:"chapter"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	var readHistory models.ReadHistory
	result := appdata.DB.Where("user_id = ? AND book = ? AND chapter = ?", user_id, req.Book, req.Chapter).First(&readHistory)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(fiber.Map{
				"error": "This book and chapter is not marked as read",
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Something went wrong, try again later",
			})
		}
	}
	appdata.DB.Delete(&readHistory)
	return c.JSON(fiber.Map{
		"message": "Chapter marked unread",
	})
}
