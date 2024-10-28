package appdata

import (
	"gorm.io/gorm"
)

var DB *gorm.DB
var SmtpServer string
var SmtpUsername string
var SmtpPassword string
var SmtpPort uint
var JwtExpiryMinutes uint
var JwtSecret []byte
var RefreshExpiryMinutes uint
var RefreshExpiryNoRemember uint
var JwtExpiryNoRemember uint
