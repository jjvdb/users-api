package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
)

func GetAvatarURL(email, name string) string {
	hash := md5.Sum([]byte(strings.ToLower(email)))
	hashHex := hex.EncodeToString(hash[:])

	fallback := fmt.Sprintf("https://ui-avatars.com/api//%s/80", name)
	fallbackURL := url.QueryEscape(fallback)

	return fmt.Sprintf("https://gravatar.com/avatar/%s?d=%s", hashHex, fallbackURL)
}
