package routes

import (
	"strings"

	"users-api/app/appdata"
	"users-api/app/models"
	"users-api/app/utils"

	"github.com/gofiber/fiber/v2"
)

// SetParallelTranslations godoc
// @Summary      Configure one or more preferred parallel translations
// @Description  Enables the current user to choose which Bible translations appear in parallel with a source translation.
// @Tags         parallel
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer JWT token" default(Bearer <token>)
// @Param        body body object true "Parallel translation configuration" example({"source_translation": "KJV", "parallel_translations": ["TOVBSI", "GOVBSI", "ASV"]})
// @Success      201  {object}  models.GenericMessage "Records created successfully"
// @Example 201 {json} {"message": "Records created successfully"}
// @Failure      400  {object}  models.ErrorResponse "Invalid request body"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /parallel [post]

func SetParallelTranslations(c *fiber.Ctx) error {

	userID := utils.GetUserFromJwt(c)

	// Deserialize JSON body
	var req struct {
		SourceTranslation    string   `json:"source_translation"`
		ParallelTranslations []string `json:"parallel_translations"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	appdata.DB.Where("user_id = ? AND translation1 = ?", userID, req.SourceTranslation).Delete(&models.ParallelTranslations{})

	// build slice of records
	pts := make([]models.ParallelTranslations, 0, len(req.ParallelTranslations))
	sourceUpper := strings.ToUpper(req.SourceTranslation)

	for _, p := range req.ParallelTranslations {
		if strings.ToUpper(p) == sourceUpper {
			continue // skip if identical ignoring case
		}
		pts = append(pts, models.ParallelTranslations{
			UserID:       userID, // set from context
			Translation1: sourceUpper,
			Translation2: strings.ToUpper(p),
		})
	}
	// bulk insert
	if err := appdata.DB.Create(&pts).Error; err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Records created successfully",
	})
}

// DeleteAllParallelTranslations godoc
// @Summary      Delete all parallel translations
// @Description  Removes all configured parallel Bible translations for the current user.
// @Tags         parallel
// @Produce      json
// @Param        Authorization header string true "Bearer JWT token" default(Bearer <token>)
// @Success      202  {object}  models.GenericMessage "Records deleted successfully"
// @Example 202 {json} {"message": "Records deleted successfully"}
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /parallel [delete]

func DeleteAllParallelTranslations(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)
	appdata.DB.Where("user_id = ?", userID).Delete(&models.ParallelTranslations{})
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Records deleted successfully",
	})
}

// DeleteParallelTranslations godoc
// @Summary      Delete a specific parallel translation
// @Description  Removes a specific parallel Bible translation for the current user.
// @Tags         parallel
// @Produce      json
// @Param        Authorization header string true "Bearer JWT token" default(Bearer <token>)
// @Param        translation path string true "Translation to remove" example("GOVBSI")
// @Success      202  {object}  models.GenericMessage "Records deleted successfully"
// @Example 202 {json} {"message": "Records deleted successfully"}
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /parallel/{translation} [delete]

func DeleteParallelTranslations(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)
	translation := c.Params("translation")
	appdata.DB.Where("user_id = ? AND translation1 = ?", userID, translation).Delete(&models.ParallelTranslations{})
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Records deleted successfully",
	})
}

// GetAllParallelTranslations godoc
// @Summary      Lists all configured parallel Bible translations
// @Description  Returns every configured set of parallel Bible translations for the current user.
// @Tags         parallel
// @Produce      json
// @Param        Authorization header string true "Bearer JWT token" default(Bearer <token>)
// @Success      200  {array}  models.ParallelTranslationResponse "List of all configured parallel Bible translations"
// @Example 200 {json} [
//   { "sourceTranslation": "KJV", "parallelTranslations": ["TOVBSI","ASV"] },
//   { "sourceTranslation": "MSLVP", "parallelTranslations": ["ASV","WEB"] }
// ]
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /parallel [get]

func GetAllParallelTranslations(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)

	var pts []models.ParallelTranslations
	if err := appdata.DB.Where("user_id = ?", userID).Find(&pts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// group by Translation1
	resultMap := make(map[string][]string)
	for _, pt := range pts {
		resultMap[pt.Translation1] = append(resultMap[pt.Translation1], pt.Translation2)
	}

	// convert map → []Req
	var response []models.ParallelTranslationResponse
	for src, parallels := range resultMap {
		response = append(response, models.ParallelTranslationResponse{
			SourceTranslation:    src,
			ParallelTranslations: parallels,
		})
	}

	return c.JSON(response)
}

// GetParallelTranslations godoc
// @Summary      Lists the parallel Bible translations for a specific source translation
// @Description  Returns the current user's configured parallel Bible translations for a given source translation.
// @Tags         parallel
// @Produce      json
// @Param        Authorization header string true "Bearer JWT token" default(Bearer <token>)
// @Param        translation path string true "Source translation code" example("KJV")
// @Success      200  {object}  models.ParallelTranslationResponse "List of parallel Bible translations for the specified source translation"
// @Example 200 {json} { "sourceTranslation": "KJV", "parallelTranslations": ["TOVBSI","ASV"] }
// @Failure      401  {object}  models.ErrorResponse "Unauthorized - missing/invalid token"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /parallel/{translation} [get]

func GetParallelTranslations(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)
	translation := c.Params("translation")

	var pts []models.ParallelTranslations
	if err := appdata.DB.Where("user_id = ? AND translation1 = ?", userID, translation).
		Find(&pts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// build response
	res := models.ParallelTranslationResponse{
		SourceTranslation:    translation,
		ParallelTranslations: make([]string, 0, len(pts)),
	}
	for _, pt := range pts {
		res.ParallelTranslations = append(res.ParallelTranslations, pt.Translation2)
	}

	return c.JSON(res)
}
