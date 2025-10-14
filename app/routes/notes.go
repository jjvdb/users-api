package routes

import (
	"strconv"
	"users-api/app/appdata"
	"users-api/app/models"
	"users-api/app/utils"

	"github.com/gofiber/fiber/v2"
)

// CreateNote godoc
// @Summary      Create a note
// @Description  Notate a specific scripture reference for the logged in user.
// @Tags         notes
// @Accept       application/x-www-form-urlencoded
// @Produce      json
// @Param        Authorization header string true "Bearer JWT token" default(Bearer <token>)
// @Param        book  formData string true "Book name, for example 'Genesis' or 'John'" example(John)
// @Param        chapter  formData int true "Chapter number (1..150)" example(15)
// @Param        verse  formData int true "Verse number (1..176)" example(2)
// @Param        note     formData string true "Note text" example(STRONGS G142: Airo = lifts up. )
// @Success      200  {object}  models.GenericMessage "Note created confirmation"
// @Failure      400  {object}  models.ErrorResponse "Invalid input or book not valid"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /note [post]

func CreateNote(c *fiber.Ctx) error {
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
	noteString := c.FormValue("note")
	if noteString == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Note empty",
		})
	}
	note := models.Note{UserID: user_id, Book: book, ChapterNumber: uint(chapterInt), VerseNumber: uint(verseNumberInt), Note: noteString}
	appdata.DB.Create(&note)
	return c.JSON(note)
}

// DeleteNote godoc
// @Summary      Delete a note
// @Description  Removes an existing note for the logged in user.
// @Tags         notes
// @Produce      json
// @Param        Authorization header string true "Bearer JWT token" default(Bearer <token>)
// @Param        noteid path int true "Unique note identifier"
// @Success      200  {object}  models.GenericMessage "Note removed confirmation"
// @Failure      400  {object}  models.ErrorResponse "Invalid noteId"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /note/{noteId} [delete]

func DeleteNote(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	noteIdString := c.Params("noteid")
	noteId, err := strconv.Atoi(noteIdString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Wrong note id",
		})
	}
	var note models.Note
	appdata.DB.First(&note, noteId)
	if note.UserID != user_id {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Note not found or it doesn't belong to you.",
		})
	}
	appdata.DB.Delete(&note)
	return c.JSON(fiber.Map{
		"message": "Note deleted successfully",
	})
}

// UpdateNote godoc
// @Summary      Update a note
// @Description  Updates an existing note for the logged in user.
// @Tags         notes
// @Accept       application/x-www-form-urlencoded
// @Produce      json
// @Param        Authorization header string true "Bearer JWT token" default(Bearer <token>)
// @Param        noteid path int true "Unique note identifier"
// @Param        note  formData string true "Updated note text" example("Every branch in me that beareth not fruit he lifts up (airo).")
// @Success      200  {object}  models.GenericMessage "Note updated confirmation"
// @Failure      400  {object}  models.ErrorResponse "Invalid noteId"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /note/{noteId} [patch]

func UpdateNote(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	noteIdString := c.Params("noteid")
	noteId, err := strconv.Atoi(noteIdString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Wrong noteId",
		})
	}
	var note models.Note
	appdata.DB.First(&note, noteId)
	if note.UserID != user_id {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Note not found or it doesn't belong to you.",
		})
	}
	noteString := c.FormValue("note")
	if noteString != "" {
		note.Note = noteString
		appdata.DB.Save(&note)
	}
	return c.JSON(note)
}

// GetNotesOfUser godoc
// @Summary      Retrieve all notes
// @Description  Retrieves all existing notes for the logged in user. Optional filters include book name, abbreviation, and chapter number.
// @Tags         notes
// @Produce      json
// @Param        Authorization header string true "Bearer JWT token" default(Bearer <token>)
// @Param        book query string false "Book name, for example 'Genesis' or 'John'" example(John)
// @Param        abbreviation query string false "Book abbreviation, for example 'Jn' or 'Gen'" example(Jn)
// @Param        chapter query int false "Chapter number (1..150)" example(3)
// @Success      200  {object}  models.Note "List of user notes"
// @Failure      400  {object}  models.ErrorResponse "Invalid query parameter"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /note [get]


func GetNotesOfUser(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	book := c.Query("book") // prefer using abbreviation (there are spaces in book names)
	abbreviation := c.Query("abbreviation")
	var notes []models.Note
	chapterString := c.Query("chapter")
	if book == "" && abbreviation == "" {
		appdata.DB.Where("user_id = ?", user_id).Find(&notes)
		return c.JSON(notes)
	}
	chapterInt, _ := strconv.Atoi(chapterString)
	if book != "" {
		if chapterInt != 0 {
			appdata.DB.Where("user_id = ? AND book = ? AND chapter_number = ?", user_id, book, uint(chapterInt)).Find(&notes)
			return c.JSON(notes)
		} else {
			appdata.DB.Where("user_id = ? AND book = ?", user_id, book).Find(&notes)
			return c.JSON(notes)
		}
	}
	if abbreviation != "" {
		for _, b := range appdata.Books {
			if b.Abbreviation == abbreviation {
				if chapterInt != 0 {
					appdata.DB.Where("user_id = ? AND book = ? AND chapter_number = ?", user_id, b.Book, uint(chapterInt)).Find(&notes)
					return c.JSON(notes)
				} else {
					appdata.DB.Where("user_id = ? AND book = ?", user_id, b.Book).Find(&notes)
					return c.JSON(notes)
				}
			}
		}
	}
	return c.JSON(notes)
}
