package routes

import (
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"time"
	"users-api/app/appdata"
	"users-api/app/models"
	"users-api/app/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CheckIfUsernameAvailable(c *fiber.Ctx) error {
	username := c.Query("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username is required",
		})
	}

	var user models.User
	result := appdata.DB.Where("username = ?", username).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(fiber.Map{
				"available": true,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	return c.JSON(fiber.Map{
		"available": false,
	})
}

func CreateUser(c *fiber.Ctx) error {
	var req models.SignupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	req.Trim()
	if utils.IsEmail(req.Username) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username cannot be an email",
		})
	}
	address, err := mail.ParseAddress(req.Email)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad Email",
		})
	}
	email := address.Address
	hashedPassword := utils.HashPassword(req.Password)
	user := models.User{Email: email, Username: req.Username, Password: hashedPassword, Name: req.Name}
	result := appdata.DB.Create(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Email or Username already exists",
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Something went wrong, try again later",
			})
		}
	}
	return c.Status(fiber.StatusCreated).JSON(user)
}

func UpdateUser(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	var user models.User
	appdata.DB.First(&user, user_id)
	email := c.FormValue("email")
	name := c.FormValue("name")
	username := c.FormValue("username")
	photoUrl := c.FormValue("photourl")
	bio := c.FormValue("bio")

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
	if username != "" {
		user.Username = username
	}
	user.Bio = bio
	user.Trim()
	result := appdata.DB.Save(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Email or Username already exists",
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Something went wrong, try again later",
			})
		}
	}
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

func ResetPassword(c *fiber.Ctx) error {
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

func ChangePassword(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	var user models.User
	appdata.DB.First(&user, user_id)
	oldPassword := c.FormValue("oldpassword")
	newPassword := c.FormValue("newPassword")
	confirmPassword := c.FormValue("confirmPassword")
	if newPassword != confirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "New password and confirm password did not match",
		})
	}
	if !utils.CheckPassword(oldPassword, user.Password) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Old password wrong. If you forgot the password, request a reset link.",
		})
	}
	hashedPassword := utils.HashPassword(newPassword)
	user.Password = hashedPassword
	appdata.DB.Save(&user)
	return c.JSON(fiber.Map{
		"message": "Password updated successfully",
	})
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
	var preference models.UserPreference
	result := appdata.DB.Where("user_id = ?", user_id).First(&preference)
	if result.Error == nil {
		user.Preference = preference
	}
	return c.JSON(user)
}

func UpdateUserPreferences(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	var userPreferences models.UserPreference
	appdata.DB.Where("user_id = ?", user_id).First(&userPreferences)
	userPreferences.UserID = user_id

	darkModeString := c.FormValue("dark_mode")
	theme := c.FormValue("theme")
	translation := c.FormValue("translation")
	lastReadBook := c.FormValue("last_read_book")
	lastReadChapterString := c.FormValue("last_read_chapter")
	fontSizeString := c.FormValue("font_size")
	fontFamilyString := c.FormValue("font_family")
	referenceAtBottom := c.FormValue("reference_at_bottom")
	copyIncludesUrl := c.FormValue("copy_includes_url")
	markAsReadAutomatically := c.FormValue("mark_as_read_automatically")
	fontSize, fontSizeError := strconv.Atoi(fontSizeString)
	fontFamily, _ := strconv.Atoi(fontFamilyString)
	lastReadChapterInt, _ := strconv.Atoi(lastReadChapterString)
	chapter := uint(lastReadChapterInt)

	switch darkModeString {
	case "true":
		userPreferences.DarkMode = true
	case "false":
		userPreferences.DarkMode = false
	}
	switch copyIncludesUrl {
	case "true":
		userPreferences.CopyIncludesUrl = true
	case "false":
		userPreferences.CopyIncludesUrl = false
	}
	switch markAsReadAutomatically {
	case "true":
		userPreferences.MarkAsReadAutomatically = true
	case "false":
		userPreferences.MarkAsReadAutomatically = false
	}
	for _, t := range appdata.AvailableTranslations {
		if t == translation {
			userPreferences.Translation = &translation
		}
	}
	if userPreferences.Translation != nil && *userPreferences.Translation != "" {
		parallelTranslations := c.FormValue("parallel_translations")
		if parallelTranslations != "" {
			userPreferences.ParallelTranslations = &parallelTranslations
		}
	}
	for _, b := range appdata.Books {
		if b.Book == lastReadBook {
			userPreferences.LastReadBook = &lastReadBook
			if chapter != 0 {
				if chapter <= b.Chapters {
					userPreferences.LastReadChapter = chapter
				}
			}
		}
	}

	if theme != "" {
		userPreferences.Theme = &theme
	}
	if fontSizeError == nil {
		userPreferences.FontSize = int(fontSize)
	}
	if fontFamily != 0 {
		userPreferences.FontFamily = uint(fontFamily)
	}
	if referenceAtBottom != "" {
		switch referenceAtBottom {
		case "true":
			userPreferences.ReferenceAtBottom = true
		case "false":
			userPreferences.ReferenceAtBottom = false
		}
	}
	appdata.DB.Save(&userPreferences)
	return c.JSON(userPreferences)
}

func DeleteUserPreferences(c *fiber.Ctx) error {
	user_id := utils.GetUserFromJwt(c)
	var userPreferences models.UserPreference
	appdata.DB.Where("user_id = ?", user_id).First(&userPreferences)
	appdata.DB.Delete(&userPreferences)
	return c.JSON(fiber.Map{
		"message": "User preferences deleted",
	})
}
