package routes

import (
	"users-api/app/appdata"
	"users-api/app/models"
	"users-api/app/utils"

	"github.com/gofiber/fiber/v2"
)

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
	for _, p := range req.ParallelTranslations {
		pts = append(pts, models.ParallelTranslations{
			UserID:       userID, // set from context
			Translation1: req.SourceTranslation,
			Translation2: p,
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

func DeleteAllParallelTranslations(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)
	appdata.DB.Where("user_id = ?", userID).Delete(&models.ParallelTranslations{})
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Records deleted successfully",
	})
}

func DeleteParallelTranslations(c *fiber.Ctx) error {
	userID := utils.GetUserFromJwt(c)
	translation := c.Params("translation")
	appdata.DB.Where("user_id = ? AND translation1 = ?", userID, translation).Delete(&models.ParallelTranslations{})
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Records deleted successfully",
	})
}

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
