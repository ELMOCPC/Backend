package httpv1

import (
	"github.com/ELMOCPC/Backend/wire"
	"github.com/gin-gonic/gin"
)

func SetupGeneralRoutes(routerGroup *gin.RouterGroup, app *wire.Application) {
	auth := routerGroup.Group("/auth")
	{
		signup := auth.Group("/signup")
		signup.Use(app.Middlewares.RateLimit.SignupRateLimit())
		{
			signup.POST("", app.Controllers.General.AuthController.SignUp)
			signup.POST("/verify", app.Controllers.General.AuthController.VerifySignUp)
		}

		login := auth.Group("/login")
		login.Use(app.Middlewares.RateLimit.LoginRateLimit())
		{
			login.POST("", app.Controllers.General.AuthController.Login)
			login.POST("/verify", app.Controllers.General.AuthController.VerifyLogin)
		}
	}

	invitations := routerGroup.Group("/invitations")
	{
		invitations.POST("/accept-team", app.Controllers.Customer.TeamController.AcceptTeamInvitation)
		invitations.POST("/reject-team", app.Controllers.Customer.TeamController.RejectTeamInvitation)
	}
}
