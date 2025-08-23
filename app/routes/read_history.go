package routes

import (
	"errors"
	"strconv"
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

func MarkBookAsRead(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)
	bookIDStr := c.Params("bookid")
	bookIDUint64, err := strconv.ParseUint(bookIDStr, 10, 64)
	if err != nil || bookIDUint64 > 66 || bookIDUint64 < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid book id",
		})
	}
	bookStruct := appdata.Books[bookIDUint64]

	// reset history in this book
	if err := appdata.DB.Where("user_id = ? AND book = ?", userID, uint(bookIDUint64)).
		Delete(&models.ReadHistory{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reset book read history",
		})
	}

	// construct read history entries
	readHistories := make([]models.ReadHistory, 0, bookStruct.Chapters)
	for ch := uint(1); ch <= bookStruct.Chapters; ch++ {
		readHistories = append(readHistories, models.ReadHistory{
			UserID:  userID,
			Book:    uint(bookIDUint64),
			Chapter: ch,
		})
	}

	// insert in batches to avoid one-by-one inserts
	if err := appdata.DB.CreateInBatches(readHistories, 100).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to mark book as read",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Book marked as read",
		"count":   len(readHistories),
	})
}

func MarkBookAsUnread(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)
	bookIDStr := c.Params("bookid")
	bookIDUint64, err := strconv.ParseUint(bookIDStr, 10, 64)
	if err != nil || bookIDUint64 > 66 || bookIDUint64 < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid book id",
		})
	}

	// Delete entries history in this book
	if err := appdata.DB.Where("user_id = ? AND book = ?", userID, uint(bookIDUint64)).
		Delete(&models.ReadHistory{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reset book read history",
		})
	}

	return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
		"message": "Read history for book deleted",
	})
}
