package routes

import (
	"errors"
	"fmt"
	"time"
	"users-api/app/appdata"
	"users-api/app/models"
	"users-api/app/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)


func LoginUser(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	var user models.User
	result := appdata.DB.Where("email = ?", req.Email).First(&user)
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
	passwordCorrect := utils.CheckPassword(req.Password, user.Password)
	if !passwordCorrect {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Wrong Password",
		})
	}
	jwtToken := utils.PrepareAccessToken(&user, req.Remember)
	refreshToken := utils.PrepareRefreshToken(&user, req.Device, req.Location, req.Remember)
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

func LogoutAll(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	appdata.DB.Where("user_id = ?", user_id).Delete(&models.RefreshToken{})
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Logout successful, it might take upto %d minutes to log out of all devices completely.", appdata.JwtExpiryMinutes)})
}

func Logout(c *fiber.Ctx) error {
	token := c.Get("Refresh")
	if token == "" {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Need refresh token in the header",
		})
	}
	var refreshToken models.RefreshToken
	result := appdata.DB.Where("token = ?", token).First(&refreshToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Refresh token invalid",
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Something went wrong, try again later",
			})
		}
	}
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Logout successful, it might take upto %d minutes to completely log out of the device completely.", appdata.JwtExpiryMinutes)})
}
