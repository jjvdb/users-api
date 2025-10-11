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

// LoginUser godoc
// @Summary      Login a user
// @Description  Authenticates a user with email and password and returns a JWT access token and a refresh token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body  models.LoginRequest  true  "User login credentials"
// @Success      200  {object}  models.LoginResponse
// @Failure      401  {object}  models.ErrorResponse
// @Router       /login [post]
func LoginUser(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid request body"})
	}
	req.Trim()
	var user models.User
	var result *gorm.DB
	if utils.IsEmail(req.EmailOrUsername) {
		result = appdata.DB.Where("email = ?", req.EmailOrUsername).First(&user)
	} else {
		result = appdata.DB.Where("username = ?", req.EmailOrUsername).First(&user)
	}
	errorMessage := ""
	var returnStatus int
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			errorMessage = "Email or Username not found"
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
	return c.Status(fiber.StatusOK).JSON(models.LoginResponse{
		AccessToken:  jwtToken,
		RefreshToken: refreshToken,
	})
}

// RefreshToken godoc
// @Summary      Refresh JWT token
// @Description  Validates the refresh token and returns a new access and refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Refresh  header  string  true  "Refresh token"
// @Success      200  {object}  map[string]string  "access_token and refresh_token"
// @Failure      400  {object}  map[string]string  "Bad request or token issues"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /refresh [post]
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
	newRefresh := utils.PrepareRefreshToken(&user, refresh.Device, refresh.Location, refresh.Remember)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token":  newJwtToken,
		"refresh_token": newRefresh,
	})
}

// LogoutAll godoc
// @Summary      Logout user from all devices
// @Description  Logs out the user from all devices by invalidating all provided refresh tokens.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true  "Bearer JWT token"
// @Success      200  {object}  map[string]string  "Logout confirmation message"
// @Failure      401  {object}  map[string]string  "Unauthorized, invalid or missing JWT"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /logout/all [post]
func LogoutAll(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	appdata.DB.Where("user_id = ?", user_id).Delete(&models.RefreshToken{})
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Logout successful, it might take upto %d minutes to log out of all devices completely.", appdata.JwtExpiryMinutes)})
}

// Logout godoc
// @Summary      Logout user from current device
// @Description  Logs out the user from all devices by invalidating all refresh tokens associated with their account.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Refresh  header  string  true  "Refresh token"
// @Success      200  {object}  map[string]string  "Logout confirmation message"
// @Failure      400  {object}  map[string]string  "Bad request or token issues"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /logout [post]
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

	// Delete the refresh token
	if err := appdata.DB.Delete(&refreshToken).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete refresh token",
		})
	}

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Logout successful, it might take up to %d minutes to log out of the device completely.", appdata.JwtExpiryMinutes),
	})
}
