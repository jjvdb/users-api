package main

import (
	"versequick-users-api/app"
)

func main() {
	server := app.NewApp()
	server.InitializeApp()
	server.InitializeDatabase()
	server.SetupRoutes()
	server.Start()
}
