package utils

import (
	"time"
	"versequick-users-api/app/appdata"
	"versequick-users-api/app/models"

	"github.com/golang-jwt/jwt/v5"
)

func PrepareAccessToken(user *models.User, expiry *time.Time) string {
	claims := jwt.MapClaims{
		"id":      user.ID,
		"expires": expiry.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(appdata.JwtSecret)
	if err == nil {
		return signedToken
	} else {
		return ""
	}
}

func PrepareRefreshToken(user *models.User, expiry *time.Time, device *string, location *string) string {
	oneRefreshPeriodBefore := time.Now().Add(-time.Duration(appdata.RefreshExpiryMinutes) * time.Minute)
	appdata.DB.Where("expires_at < ?", oneRefreshPeriodBefore).Delete(&models.RefreshToken{})
	tokenString := GenerateAlphanumeric(25)
	refreshToken := models.RefreshToken{UserID: user.ID, Device: *device, Location: location, Token: tokenString, ExpiresAt: *expiry}
	result := appdata.DB.Create(&refreshToken)
	if result.Error == nil {
		return tokenString
	} else {
		return ""
	}
}
