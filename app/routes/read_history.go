package routes

import (
	"errors"
	"strconv"
	"strings"
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
		BookID       uint   `json:"book_id"`
		Book         string `json:"book"`
		Abbreviation string `json:"abbreviation"`
		Chapter      uint   `json:"chapter"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	req.Abbreviation = strings.ToUpper(req.Abbreviation)

	var bookNum uint

	if req.BookID != 0 {
		if req.BookID < 1 || req.BookID > 66 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid book",
			})
		}
		bookNum = req.BookID

	} else if req.Book != "" {
		for i := range appdata.Books {
			if appdata.Books[i].Book == req.Book {
				bookNum = uint(i + 1)
				break
			}
		}
	} else if req.Abbreviation != "" {
		for i := range appdata.Books {
			if appdata.Books[i].Abbreviation == req.Abbreviation {
				bookNum = uint(i + 1)
				break
			}
		}
	}

	if bookNum == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid book",
		})
	}

	bookStruct := appdata.Books[bookNum-1]

	if bookStruct.Chapters == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chapter",
		})
	}

	if req.Chapter > bookStruct.Chapters {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chapter number",
		})
	}
	readHistory := models.ReadHistory{UserID: userID, Book: bookNum, Chapter: req.Chapter}
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
		"book":         bookStruct.Book,
		"abbreviation": bookStruct.Abbreviation,
		"chapter":      req.Chapter,
		"message":      "Chapter marked as read",
	})
}

func MarkChapterAsUnread(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)

	// Deserialize JSON body
	var req struct {
		BookID       uint   `json:"book_id"`
		Book         string `json:"book"`
		Abbreviation string `json:"abbreviation"`
		Chapter      uint   `json:"chapter"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	req.Abbreviation = strings.ToUpper(req.Abbreviation)

	var bookNum uint

	if req.BookID != 0 {
		if req.BookID < 1 || req.BookID > 66 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid book",
			})
		}
		bookNum = req.BookID

	} else if req.Book != "" {
		for i := range appdata.Books {
			if appdata.Books[i].Book == req.Book {
				bookNum = uint(i + 1)
				break
			}
		}
	} else if req.Abbreviation != "" {
		for i := range appdata.Books {
			if appdata.Books[i].Abbreviation == req.Abbreviation {
				bookNum = uint(i + 1)
				break
			}
		}
	}

	if bookNum == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid book",
		})
	}

	var readHistory models.ReadHistory
	result := appdata.DB.Where("user_id = ? AND book = ? AND chapter = ?", user_id, bookNum, req.Chapter).First(&readHistory)
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
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"book":         appdata.Books[bookNum-1].Book,
		"abbreviation": appdata.Books[bookNum-1].Abbreviation,
		"chapter":      req.Chapter,
		"message":      "Chapter marked as unread",
	})
}

func MarkBookAsRead(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)

	bookIDStr := c.Params("bookid")
	var bookIDUint64 uint64
	bookIDUint64, err := strconv.ParseUint(bookIDStr, 10, 64)
	if err != nil {
		if bookIDUint64 > 66 || bookIDUint64 < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid book id",
			})
		}
	} else {
		for i := range appdata.Books {
			if strings.ToUpper(bookIDStr) == appdata.Books[i].Abbreviation {
				bookIDUint64 = uint64(i + 1)
				break
			}
		}
		if bookIDUint64 == 0 {
			for i := range appdata.Books {
				if strings.EqualFold(strings.ReplaceAll(bookIDStr, "-", " "), appdata.Books[i].Book) {
					bookIDUint64 = uint64(i + 1)
					break
				}
			}
		}
	}

	if bookIDUint64 == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid book id",
		})
	}

	bookStruct := appdata.Books[bookIDUint64-1]

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
		"message":      "Book marked as read",
		"book":         bookStruct.Book,
		"abbreviation": bookStruct.Abbreviation,
		"count":        len(readHistories),
	})
}

func MarkBookAsUnread(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)
	bookIDStr := c.Params("bookid")
	var bookIDUint64 uint64
	bookIDUint64, err := strconv.ParseUint(bookIDStr, 10, 64)
	if err != nil {
		if bookIDUint64 > 66 || bookIDUint64 < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid book id",
			})
		}
	} else {
		for i := range appdata.Books {
			if strings.ToUpper(bookIDStr) == appdata.Books[i].Abbreviation {
				bookIDUint64 = uint64(i + 1)
				break
			}
		}
		if bookIDUint64 == 0 {
			for i := range appdata.Books {
				if strings.EqualFold(strings.ReplaceAll(bookIDStr, "-", " "), appdata.Books[i].Book) {
					bookIDUint64 = uint64(i + 1)
					break
				}
			}
		}
	}

	if bookIDUint64 == 0 {
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

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"book":         appdata.Books[bookIDUint64-1].Book,
		"abbreviation": appdata.Books[bookIDUint64-1].Abbreviation,
		"message":      "Read history for book deleted",
	})
}

func GetReadChaptersOfBook(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)

	// Parse book ID from route param
	bookIDStr := c.Params("bookid")
	bookIDUint64, err := strconv.ParseUint(bookIDStr, 10, 64)
	if err != nil || bookIDUint64 < 1 || bookIDUint64 > 66 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid book id",
		})
	}
	bookID := uint(bookIDUint64)

	// Query DB for read chapters
	var histories []models.ReadHistory
	if err := appdata.DB.
		Where("user_id = ? AND book = ?", userID, bookID).
		Find(&histories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch read chapters",
		})
	}

	// Collect chapters
	readChapters := make([]uint, 0, len(histories))
	for _, h := range histories {
		readChapters = append(readChapters, h.Chapter)
	}

	// Return JSON
	return c.JSON(fiber.Map{
		"book_id":       bookID,
		"book":          appdata.Books[bookID-1].Book,
		"abbreviation":  appdata.Books[bookID-1].Abbreviation,
		"read_chapters": readChapters,
	})
}

func GetReadBooksStatus(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)

	// Fetch all read history for the user
	var histories []models.ReadHistory
	if err := appdata.DB.
		Where("user_id = ?", userID).
		Find(&histories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch read history",
		})
	}

	// Of the form {"book_id": count}, like {1: 10}, which means Genesis 10 chapters in read history
	readHistories := make(map[uint]uint)

	for i := range histories {
		readHistories[histories[i].Book] += 1
	}
	result := make([]models.ReadBook, 0, 66)
	for i := range appdata.Books {
		result = append(result, models.ReadBook{Book: appdata.Books[i].Book, Abbreviation: appdata.Books[i].Abbreviation, Status: models.StatusNotStarted})
	}
	for bookID, count := range readHistories {
		if count == appdata.Books[bookID-1].Chapters {
			result[bookID-1].Status = models.StatusComplete
		} else if count > 0 {
			result[bookID-1].Status = models.StatusPartial
		}
	}
	return c.Status(fiber.StatusOK).JSON(result)
}
