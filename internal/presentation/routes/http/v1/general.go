package httpv1

import (
	"github.com/ELMOCPC/Backend/wire"
	"github.com/gin-gonic/gin"
)

func SetupGeneralRoutes(routerGroup *gin.RouterGroup, app *wire.Application) {
	auth := routerGroup.Group("/auth")
	{
		auth.POST("/signup", app.Controllers.General.AuthController.SignUp)
		auth.POST("/signup/verify", app.Controllers.General.AuthController.VerifySignUp)

		auth.POST("/login", app.Controllers.General.AuthController.Login)
		auth.POST("/login/verify", app.Controllers.General.AuthController.VerifyLogin)
	}

	invitations := routerGroup.Group("/invitations")
	{
		invitations.POST("/accept-team", app.Controllers.Customer.TeamController.AcceptTeamInvitation)
		invitations.POST("/reject-team", app.Controllers.Customer.TeamController.RejectTeamInvitation)
	}
}
