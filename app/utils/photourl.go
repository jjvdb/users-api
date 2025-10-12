package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
)

// GetAvatarURL returns the Gravatar URL with UI Avatars fallback.
func GetAvatarURL(email, name string) string {
	// Compute MD5 hash of the lowercase email
	hash := md5.Sum([]byte(strings.ToLower(email)))
	hashHex := hex.EncodeToString(hash[:])

	// Encode fallback name for UI Avatars
	fallback := url.QueryEscape(name)
	fallbackURL := fmt.Sprintf("https://ui-avatars.com/api/?name=%s", fallback)

	// Combine into final Gravatar URL
	return fmt.Sprintf("https://gravatar.com/avatar/%s?d=%s", hashHex, fallbackURL)
}
