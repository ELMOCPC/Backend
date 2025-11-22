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
		&entity.Team{},
		&entity.TeamInvitation{},
		&entity.OTP{},
		&entity.TeamInvitationToken{},
		&entity.PaymentReceipt{},
	)

	app.Seeds.RoleSeeder.SeedRoles()

	routes.Run(ginEngine, app)

	ginEngine.Run(fmt.Sprintf(":%v", config.Env.Server.Port))
}
