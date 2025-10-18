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

// MarkChapterAsRead godoc
// @Summary      Record a completed Bible chapter
// @Description  Records that the current user has finished reading a specific Bible chapter.
// @Tags         read_history
// @Accept       json
// @Produce      json
// @Param        chapter  body  models.BibleChapter  true  "Details which Bible chapter to mark as completed"
// @Success      201  {object}  models.MarkChapterAsReadResponse "Chapter successfully recorded as read"
// @Example 201 {json} {
//   "message": "Chapter marked as read",
//   "book": "John",
//   "abbreviation": "JHN",
//   "chapter": 3
// }
// @Failure      400  {object}  models.ErrorResponse "Invalid input"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /markchapterasread [post]

func MarkChapterAsRead(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)

	// Deserialize JSON body
	var req models.BibleChapter
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewInvalidRequestBodyError())
	}

	req.Abbreviation = strings.ToUpper(req.Abbreviation)

	var bookNum uint

	if req.BookID != 0 {
		if req.BookID < 1 || req.BookID > 66 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid book"})
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
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid book"})
	}

	bookStruct := appdata.Books[bookNum-1]

	if bookStruct.Chapters == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid chapter"})
	}

	if req.Chapter > bookStruct.Chapters {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid chapter number"})
	}
	readHistory := models.ReadHistory{UserID: userID, Book: bookNum, Chapter: req.Chapter}
	result := appdata.DB.Create(&readHistory)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			// Successful response is intentional
			return c.JSON(models.ErrorResponse{Error: "This chapter is already marked as read"})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(models.NewInternalError())
		}
	}
	response := models.MarkChapterAsReadResponse{
		Book:         bookStruct.Book,
		Abbreviation: bookStruct.Abbreviation,
		Chapter:      req.Chapter,
		Message:      "Chapter marked as read",
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// MarkChapterAsUnread godoc
// @Summary      Unmark a completed Bible chapter
// @Description  Removes a Bible chapter from the current user's list of completed (read) chapters.
// @Tags         read_history
// @Accept       json
// @Produce      json
// @Param        chapter  body  models.BibleChapter  true  "Details which Bible chapter to mark as unread"
// @Success      200  {object}  models.MarkChapterAsReadResponse "Chapter successfully marked as unread"
// @Example 200 {json} {
//   "message": "Chapter marked as unread",
//   "book": "John",
//   "abbreviation": "JHN",
//   "chapter": 3
// }
// @Failure      400  {object}  models.ErrorResponse "Invalid input"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /markchapterasread [delete]

func MarkChapterAsUnread(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)

	// Deserialize JSON body
	var req models.BibleChapter
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewInvalidRequestBodyError())
	}

	req.Abbreviation = strings.ToUpper(req.Abbreviation)

	var bookNum uint

	if req.BookID != 0 {
		if req.BookID < 1 || req.BookID > 66 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid book"})
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
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid book"})
	}

	var readHistory models.ReadHistory
	result := appdata.DB.Where("user_id = ? AND book = ? AND chapter = ?", user_id, bookNum, req.Chapter).First(&readHistory)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(models.ErrorResponse{Error: "This chapter is not marked as read"})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(models.NewInternalError())
		}
	}
	appdata.DB.Delete(&readHistory)

	response := models.MarkChapterAsReadResponse{
		Book:         appdata.Books[bookNum-1].Book,
		Abbreviation: appdata.Books[bookNum-1].Abbreviation,
		Chapter:      req.Chapter,
		Message:      "Chapter marked as unread",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// MarkBookAsRead godoc
// @Summary      Mark a fully completed Bible book
// @Description  Records that the current user has finished reading all the chapters in a specific Bible book.
// @Tags         read_history
// @Accept       json
// @Produce      json
// @Param        bookid   path  string  true  "Book identifier, either: the full name of the book (for example, '3 John'), abbreviation (for example, '3JN'), or the book number (1-66)"
// @Success      201  {object}  models.MarkBookReadResponse "Book successfully marked as read"
// @Example 201 {json} {
//   "message": "Book marked as read",
//   "book": "John",
//   "abbreviation": "JHN",
//   "count": 21
// }
// @Failure      400  {object}  models.ErrorResponse "Invalid input"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router 		 /markbookasread/{bookid} [post]

func MarkBookAsRead(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)

	bookIDStr := c.Params("bookid")
	var bookIDUint64 uint64
	bookIDUint64, err := strconv.ParseUint(bookIDStr, 10, 64)
	if err == nil {
		if bookIDUint64 > 66 || bookIDUint64 < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid book id"})
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
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid book id"})
	}

	bookStruct := appdata.Books[bookIDUint64-1]

	// reset history in this book
	if err := appdata.DB.Where("user_id = ? AND book = ?", userID, uint(bookIDUint64)).
		Delete(&models.ReadHistory{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "Invalid book id"})
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
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewInternalError())
	}

	response := models.MarkBookReadResponse{
		Message:      "Book marked as read",
		Book:         bookStruct.Book,
		Abbreviation: bookStruct.Abbreviation,
		Count:        len(readHistories),
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// MarkBookAsUnread godoc
// @Summary      Unmark a fully completed Bible book
// @Description  Removes all Bible chapters of a specific book from the current user's list of completed (read) chapters.
// @Tags         read_history
// @Accept       json
// @Produce      json
// @Param        bookid   path  string  true  "Book identifier, either: the full name of the book (for example, '3 John'), abbreviation (for example, '3JN'), or the book number (1-66)"
// @Success      200  {object}  models.MarkBookReadResponse "Book successfully marked as unread"
// @Example 200 {json} {
//   "message": "Book marked as unread",
//   "book": "John",
//   "abbreviation": "JHN",
//   "count": 21
// }
// @Failure      400  {object}  models.ErrorResponse "Invalid input"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router 		 /markbookasunread/{bookid} [delete]

func MarkBookAsUnread(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)
	bookIDStr := c.Params("bookid")
	var bookIDUint64 uint64
	bookIDUint64, err := strconv.ParseUint(bookIDStr, 10, 64)
	if err == nil {
		if bookIDUint64 > 66 || bookIDUint64 < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid book id"})
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
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid book id"})
	}

	// Delete entries history in this book
	if err := appdata.DB.Where("user_id = ? AND book = ?", userID, uint(bookIDUint64)).
		Delete(&models.ReadHistory{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewInternalError())
	}

	response := models.MarkBookReadResponse{
		Message:      "Book marked as unread",
		Book:         appdata.Books[bookIDUint64-1].Book,
		Abbreviation: appdata.Books[bookIDUint64-1].Abbreviation,
		Count:        int(appdata.Books[bookIDUint64-1].Chapters), // Not counting the number of rows deleted from the DB.
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// GetReadChaptersOfBook godoc
// @Summary      Get read chapter numbers for a Bible book
// @Description  Retrieves a list of chapter numbers that the current user has marked as read for the specified Bible book.
// @Tags         read_history
// @Accept       json
// @Produce      json
// @Param        bookid   path  string  true  "Book identifier, either: the full name of the book (for example, '3 John'), abbreviation (for example, '3JN'), or the book number (1-66)"
// @Success      200  {object}  models.BookReadChaptersResponse "List of read chapters for the specified book"
// @Example 200 {json} {
//   "bookId": 43,
//   "book": "John",
//   "abbreviation": "JHN",
//   "readChapters": [1,2,3,4,5]
// }
// @Failure      400  {object}  models.ErrorResponse "Invalid book ID"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router 		 /readchaptersofbook/{bookid} [get]

func GetReadChaptersOfBook(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)

	bookIDStr := c.Params("bookid")
	var bookIDUint64 uint64
	bookIDUint64, err := strconv.ParseUint(bookIDStr, 10, 64)
	if err == nil {
		if bookIDUint64 > 66 || bookIDUint64 < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid book id"})
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
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid book id"})
	}
	bookID := uint(bookIDUint64)

	// Query DB for read chapters
	var histories []models.ReadHistory
	if err := appdata.DB.
		Where("user_id = ? AND book = ?", userID, bookID).
		Find(&histories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewInternalError())
	}

	// Collect chapters
	readChapters := make([]uint, 0, len(histories))
	for _, h := range histories {
		readChapters = append(readChapters, h.Chapter)
	}

	response := models.BookReadChaptersResponse{
		BookID:       bookID,
		Book:         appdata.Books[bookID-1].Book,
		Abbreviation: appdata.Books[bookID-1].Abbreviation,
		ReadChapters: readChapters,
	}

	return c.JSON(response)

}

// GetReadBooksStatus godoc
// @Summary      Get reading progress for all Bible books
// @Description  Retrieves the reading status for each book in the Bible for the current user, indicating whether it is complete, partially read, or not started.
// @Tags         read_history
// @Accept       json
// @Produce      json
// @Success      200  {array}  models.ReadBook "List of books with reading status"
// @Example 200 {json} [
//   { "book": "Genesis", "abbreviation": "GEN", "status": "complete" },
//   { "book": "Exodus", "abbreviation": "EXO", "status": "partial" },
//   { "book": "Leviticus", "abbreviation": "LEV", "status": "not_started" }
// ]
// @Failure      401  {object} models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object} models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /readbooksstatus [get]

func GetReadBooksStatus(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)

	// Fetch all read history for the user
	var histories []models.ReadHistory
	if err := appdata.DB.
		Where("user_id = ?", userID).
		Find(&histories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewInternalError())
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
