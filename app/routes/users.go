package routes

import (
	"errors"
	"fmt"
	"net/mail"
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
	address, err := mail.ParseAddress(email)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad Email",
		})
	}
	email = address.Address
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

func UpdateUser(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	var user models.User
	appdata.DB.First(&user, user_id)
	email := c.FormValue("email")
	name := c.FormValue("name")
	photoUrl := c.FormValue("photourl")

	if email != "" {
		address, err := mail.ParseAddress(email)
		if err == nil {
			user.Email = address.Address
			user.IsActivated = false
		}
	}
	if name != "" {
		user.Name = name
	}
	if photoUrl != "" {
		user.PhotoUrl = &photoUrl
	}
	appdata.DB.Save(&user)
	return c.JSON(user)
}

func SendForgotPasswordEmail(c *fiber.Ctx) error {
	email := c.FormValue("email")
	var user models.User
	result := appdata.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Email not found in database",
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Something went wrong, try again later",
			})
		}
	}
	randString := utils.GenerateAlphanumeric(25)
	now := time.Now()
	expiresAt := now.Add(time.Duration(appdata.ResetValidMinutes) * time.Minute)
	forgotPassword := models.ForgotPassword{UserID: user.ID, Token: randString, ExpiresAt: expiresAt}
	appdata.DB.Where("expires_at < ?", now).Delete(&models.ForgotPassword{})
	appdata.DB.Where("user_id = ?", user.ID).Delete(&models.ForgotPassword{})
	appdata.DB.Create(&forgotPassword)
	resetLink := fmt.Sprintf("https://versequick.com/changepassword/%s", randString)
	emailBody := fmt.Sprintf("Visit <a href=%s>%s</a> to change your password", resetLink, resetLink)
	err := utils.SendEmail(email, "Reset your password", emailBody, true)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Something went wrong, try again later",
		})
	} else {
		return c.JSON(fiber.Map{"message": "Password reset email sent successfully"})
	}
}

func ChangePassword(c *fiber.Ctx) error {
	token := c.FormValue("token")
	var forgotPassword models.ForgotPassword
	result := appdata.DB.Where("token = ?", token).First(&forgotPassword)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password reset token invalid, get a new one at /changepassword",
		})
	}
	if forgotPassword.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password reset token expired, get a new one at /changepassword",
		})
	}
	newPassword := c.FormValue("password")
	var user models.User
	appdata.DB.First(&user, forgotPassword.UserID)
	hashedPassword := utils.HashPassword(newPassword)
	user.Password = hashedPassword
	appdata.DB.Save(&user)
	appdata.DB.Delete(&forgotPassword)
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Password changed successfully. The link is valid for %d minutes.", appdata.ResetValidMinutes)})
}

func SendEmailVerificationEmail(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	var user models.User
	result := appdata.DB.First(&user, user_id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot find user in database",
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Something went wrong, try again later",
			})
		}
	}
	if user.IsActivated {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email already verified",
		})
	}
	token := utils.GenerateAlphanumeric(25)
	now := time.Now()
	expiresAt := now.Add(time.Duration(appdata.ResetValidMinutes) * time.Minute)
	verifyEmail := models.VerifyEmail{UserID: user_id, Token: token, ExpiresAt: expiresAt}
	appdata.DB.Where("expires_at < ?", now).Delete(&models.VerifyEmail{})
	appdata.DB.Where("user_id = ?", user.ID).Delete(&models.VerifyEmail{})
	appdata.DB.Create(&verifyEmail)
	verifyLink := fmt.Sprintf("https://versequick.com/verifyemail/%s", token)
	emailBody := fmt.Sprintf("Click the link below to verify your email.<br><br><a href=%s>%s</a>", verifyLink, verifyLink)
	err := utils.SendEmail(user.Email, "Verify your email", emailBody, true)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Something went wrong, try again later",
		})
	} else {
		return c.JSON(fiber.Map{"message": fmt.Sprintf("Verification email sent successfully. The link is valid for %d minutes.", appdata.ResetValidMinutes)})
	}
}

func VerifyEmail(c *fiber.Ctx) error {
	token := c.FormValue("token")
	var verifyEmail models.VerifyEmail
	result := appdata.DB.Where("token = ?", token).First(&verifyEmail)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Verification token invalid, get a new one at /sendemailverificationemail",
		})
	}
	if verifyEmail.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Verification token expired, get a new one at /sendemailverificationemail",
		})
	}
	var user models.User
	appdata.DB.First(&user, verifyEmail.UserID)
	user.IsActivated = true
	appdata.DB.Save(&user)
	appdata.DB.Delete(&verifyEmail)
	return c.JSON(fiber.Map{"message": "Email verified successfully"})
}

func GetSelfInfo(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	var user models.User
	appdata.DB.First(&user, user_id)
	return c.JSON(user)
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
