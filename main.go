package main

import (
	"fmt"

	"github.com/ELMOCPC/Backend/bootstrap"
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/presentation/routes"
	"github.com/ELMOCPC/Backend/wire"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.DisableConsoleColor()
	ginEngine := gin.New()

	config := bootstrap.Run()

	app, err := wire.InitializeApplication(config)
	if err != nil {
		panic(err)
	}

	app.Database.DB.GetDB().AutoMigrate(
		&entity.User{},
		&entity.Role{},
		&entity.Permission{},
	)

	// Run database migrations and seeders
	app.Seeds.RoleSeeder.SeedRoles()

	// Setup and run routes
	routes.Run(ginEngine, app)

	// Start the server
	ginEngine.Run(fmt.Sprintf(":%v", config.Env.Server.Port))
}
