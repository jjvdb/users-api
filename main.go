// @title           Scripture users-api
// @version         1.0
// @description     Fiber API with Swagger
// @host            api.scripture.pp.ua
// @BasePath        /users

package main

import (
	"users-api/app"
)

func main() {
	server := app.NewApp()
	server.InitializeApp()
	server.InitializeDatabase()
	server.SetupRoutes()
	server.Start()
}
