package app

import (
	"log"
	"os"
	"strconv"
	"versequick-users-api/app/appdata"
	"versequick-users-api/app/models"
	"versequick-users-api/app/routes"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type App struct {
	Fiber *fiber.App
}

func NewApp() *App {
	fiberApp := fiber.New(fiber.Config{
		AppName: "VerseQuick Users API",
	})
	return &App{
		Fiber: fiberApp,
	}
}

func getExpiryMinutes(s string) uint {
	value := os.Getenv(s)
	if value == "" {
		log.Fatal(s + " is not set")
	}
	valueUint, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		log.Fatal(s + " could not be parsed as minutes")
	}
	return uint(valueUint)
}

func (app *App) InitializeApp() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load environment variables from .env file")
	}
	appdata.JwtExpiryMinutes = getExpiryMinutes("JWT_EXPIRY_MINUTES")
	appdata.RefreshExpiryMinutes = getExpiryMinutes("REFRESH_EXPIRY_MINUTES")
	appdata.RefreshExpiryNoRemember = getExpiryMinutes("REFRESH_EXPIRY_NO_REMEMBER")
	appdata.JwtExpiryNoRemember = getExpiryMinutes("JWT_EXPIRY_NO_REMEMBER")
	appdata.SmtpServer = os.Getenv("SMTP_SERVER")
	appdata.SmtpPassword = os.Getenv("SMTP_PASSWORD")
	envSmtpPort, _ := strconv.ParseUint(os.Getenv("SMTP_PORT"), 10, 32)
	appdata.SmtpPort = uint(envSmtpPort)
	appdata.SmtpUsername = os.Getenv("SMTP_FROM")
	if appdata.SmtpPort == 0 || appdata.SmtpServer == "" || appdata.SmtpPassword == "" || appdata.SmtpUsername == "" {
		log.Fatal("Failed to load environment variables for SMTP settings.")
	}
	jwtSecretString := os.Getenv("JWT_SECRET")
	if jwtSecretString == "" {
		log.Fatal("Failed to load JWT_SECRET from .env")
	}
	appdata.JwtSecret = []byte(jwtSecretString)
}

func (app *App) InitializeDatabase() {
	var err error
	dsn := os.Getenv("DSN")
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			IgnoreRecordNotFoundError: true,
		},
	)
	appdata.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         gormLogger,
	})
	if err != nil {
		log.Fatal("Failed to connect to database")
	}
	modelsToMigrate := []interface{}{
		&models.User{},
		&models.RefreshToken{},
	}
	for _, model := range modelsToMigrate {
		if err := appdata.DB.AutoMigrate(model); err != nil {
			log.Fatal("Failed to do database migrations")
		}
	}
}

func (app *App) SetupRoutes() {
	app.Fiber.Get("/", routes.Home)
	app.Fiber.Post("/users", routes.CreateUser)
	app.Fiber.Post("/login", routes.LoginUser)
	app.Fiber.Post("/refreshtoken", routes.RefreshToken)

	app.Fiber.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: appdata.JwtSecret},
	}))

	app.Fiber.Put("/users", routes.UpdateUser)
}

func (app *App) Start() {
	hostUrl := os.Getenv("HOST_URL")
	log.Fatal(app.Fiber.Listen(hostUrl))
}
