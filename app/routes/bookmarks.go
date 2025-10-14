package routes

import (
	"strconv"
	"users-api/app/appdata"
	"users-api/app/models"
	"users-api/app/utils"

	"github.com/gofiber/fiber/v2"
)

// AddBookmark godoc
// @Summary      Add a bookmark.
// @Description  Marks a specific scripture reference for the logged in user.
// @Tags         bookmarks
// @Accept       application/x-www-form-urlencoded
// @Produce      json
// @Param        Authorization header string true "Bearer JWT token" default(Bearer <token>)
// @Param        book  formData string true "Book name, for example 'Genesis' or 'John'" example(John)
// @Param        chapter  formData int true "Chapter number (1..150)" example(15)
// @Param        verse  formData int true "Verse number (1..176)" example(5)
// @Success      200  {object}  models.GenericMessage "Bookmark created confirmation"
// @Failure      400  {object}  models.ErrorResponse "Invalid input or book not valid"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /bookmark [post]

func AddBookmark(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	book := c.FormValue("book")
	var bookStruct appdata.Book
	for _, b := range appdata.Books {
		if b.Book == book {
			bookStruct = b
			break
		}
	}
	if bookStruct.Book == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Book not valid",
		})
	}
	chapterString := c.FormValue("chapter")
	verseNumberString := c.FormValue("verse")
	chapterInt, err := strconv.Atoi(chapterString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Chapter number not a valid number",
		})
	}
	verseNumberInt, err := strconv.Atoi(verseNumberString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Verse Number not a valid number",
		})
	}
	var bookmark models.Bookmark = models.Bookmark{UserID: user_id, Book: book, ChapterNumber: uint(chapterInt), VerseNumber: uint(verseNumberInt)}
	appdata.DB.Create(&bookmark)
	return c.JSON(fiber.Map{
		"message": "Created bookmark",
	})
}

// DeleteBookmark godoc
// @Summary      Delete a bookmark.
// @Description  Removes an existing bookmark.
// @Tags         bookmarks
// @Accept       application/x-www-form-urlencoded
// @Produce      json
// @Param        Authorization header string true "Bearer JWT token" default(Bearer <token>)
// @Param        book  formData string true "Book name, for example 'Genesis' or 'John'" example(John)
// @Param        chapter  formData int true "Chapter number (1..150)" example(15)
// @Param        verse  formData int true "Verse number (1..176)" example(5)
// @Success      200  {object}  models.GenericMessage "Bookmark removed confirmation"
// @Failure      400  {object}  models.ErrorResponse "Invalid input or book not valid"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /bookmark [delete]

func DeleteBookmark(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	book := c.FormValue("book")
	chapterString := c.FormValue("chapter")
	verseNumberString := c.FormValue("verse")
	chapterInt, err := strconv.Atoi(chapterString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Chapter number not a valid number",
		})
	}
	verseNumberInt, err := strconv.Atoi(verseNumberString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Verse Number not a valid number",
		})
	}
	var bookmark models.Bookmark
	appdata.DB.Where("user_id = ? AND book = ? AND chapter_number = ? and verse_number = ?", user_id, book, uint(chapterInt), uint(verseNumberInt)).First(&bookmark)
	appdata.DB.Delete(&bookmark)
	return c.JSON(fiber.Map{
		"message": "Bookmark deleted",
	})
}
