package routes

import (
	"errors"
	"time"
	"versequick-users-api/app/appdata"
	"versequick-users-api/app/models"
	"versequick-users-api/app/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CreateUser(c *fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	name := c.FormValue("name")
	if email == "" || password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}
	hashedPassword := utils.HashPassword(password)
	user := models.User{Email: email, Password: hashedPassword, Name: name}
	result := appdata.DB.Create(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Email already exists",
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Something went wrong, try again later",
			})
		}
	}
	return c.Status(fiber.StatusCreated).JSON(user)
}

func LoginUser(c *fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	remember := c.FormValue("remember") == "true"
	device := c.FormValue("device")
	location := c.FormValue("location")

	var user models.User
	result := appdata.DB.Where("email = ?", email).First(&user)
	errorMessage := ""
	var returnStatus int
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			errorMessage = "Email not found"
			returnStatus = fiber.StatusNotFound
		} else {
			errorMessage = "Something went wrong, try again later"
			returnStatus = fiber.StatusInternalServerError
		}
		return c.Status(returnStatus).JSON(fiber.Map{
			"error": errorMessage,
		})
	}
	passwordCorrect := utils.CheckPassword(password, user.Password)
	if !passwordCorrect {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Wrong Password",
		})
	}
	jwtToken := utils.PrepareAccessToken(&user, remember)
	refreshToken := utils.PrepareRefreshToken(&user, &device, &location, remember)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token":  jwtToken,
		"refresh_token": refreshToken,
	})
}

func RefreshToken(c *fiber.Ctx) error {
	token := c.Get("Refresh")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad Token",
		})
	}
	var refresh models.RefreshToken
	result := appdata.DB.Where("token = ?", token).First(&refresh)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Token not found",
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Something went wrong, try again later",
			})
		}
	}
	now := time.Now()
	if refresh.ExpiresAt.Before(now) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Token expired, get a new one at /login",
		})
	}
	if refresh.Revoked {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Token reuse detected, get a new one at /login",
		})
	}
	refresh.Revoked = true
	appdata.DB.Save(&refresh)
	var user models.User
	appdata.DB.First(&user, refresh.UserID)
	newJwtToken := utils.PrepareAccessToken(&user, refresh.Remember)
	newRefresh := utils.PrepareRefreshToken(&user, &refresh.Device, refresh.Location, refresh.Remember)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token":  newJwtToken,
		"refresh_token": newRefresh,
	})
}
