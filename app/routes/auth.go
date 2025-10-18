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
		return c.Status(returnStatus).JSON(models.ErrorResponse{Error: errorMessage})
	}
	passwordCorrect := utils.CheckPassword(req.Password, user.Password)
	if !passwordCorrect {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Wrong password"})
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
// @Success      200  {object}  models.LoginResponse
// @Failure      401  {object}  models.ErrorResponse
// @Router       /refresh [post]
func RefreshToken(c *fiber.Ctx) error {
	token := c.Get("Refresh")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "No refresh token in header. You'll have to send the previous refresh token to get a new set now"})
	}
	var refresh models.RefreshToken
	result := appdata.DB.Where("token = ?", token).First(&refresh)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Refresh token not found in DB"})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(models.NewInternalError())
		}
	}
	now := time.Now()
	if refresh.ExpiresAt.Before(now) {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Refresh token expired, login again."})
	}
	if refresh.Revoked {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Refresh token reuse detected, login again."})
	}
	refresh.Revoked = true
	appdata.DB.Save(&refresh)
	var user models.User
	appdata.DB.First(&user, refresh.UserID)
	newJwtToken := utils.PrepareAccessToken(&user, refresh.Remember)
	newRefresh := utils.PrepareRefreshToken(&user, refresh.Device, refresh.Location, refresh.Remember)
	return c.Status(fiber.StatusOK).JSON(models.LoginResponse{AccessToken: newJwtToken, RefreshToken: newRefresh})
}

// LogoutAll godoc
// @Summary      Logout user from all devices
// @Description  Logs out the user from all devices by invalidating all provided refresh tokens.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"
// @Success      200  {object}  models.GenericMessage
// @Failure      401  {object}  models.ErrorResponse
// @Router       /logout/all [post]
func LogoutAll(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	appdata.DB.Where("user_id = ?", user_id).Delete(&models.RefreshToken{})
	return c.JSON(models.GenericMessage{Message: fmt.Sprintf("Logout successful, it might take upto %d minutes to log out of all devices completely.", appdata.JwtExpiryMinutes)})
}

// Logout godoc
// @Summary      Logout user from current device
// @Description  Logs out the user from all devices by invalidating all refresh tokens associated with their account.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Refresh  header  string  true  "Refresh token"
// @Success      200  {object}  models.GenericMessage
// @Failure      400  {object}  models.ErrorResponse
// @Router       /logout [post]
func Logout(c *fiber.Ctx) error {
	token := c.Get("Refresh")
	if token == "" {
		c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Need refresh token in the header"})
	}
	var refreshToken models.RefreshToken
	result := appdata.DB.Where("token = ?", token).First(&refreshToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{Error: "Refresh token not found in DB"})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(models.NewInternalError())
		}
	}

	// Delete the refresh token
	if err := appdata.DB.Delete(&refreshToken).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewInternalError())
	}

	return c.JSON(models.GenericMessage{
		Message: fmt.Sprintf("Logout successful, it might take up to %d minutes to log out of the device completely.", appdata.JwtExpiryMinutes),
	})
}
