package httpv1

import (
	"github.com/ELMOCPC/Backend/wire"
	"github.com/gin-gonic/gin"
)

func SetupGeneralRoutes(routerGroup *gin.RouterGroup, app *wire.Application) {
	auth := routerGroup.Group("/auth")
	{
		// Signup flow
		auth.POST("/signup", app.Controllers.General.AuthController.SignUp)
		auth.POST("/signup/verify", app.Controllers.General.AuthController.VerifySignUp)

		// Login flow
		auth.POST("/login", app.Controllers.General.AuthController.Login)
		auth.POST("/login/verify", app.Controllers.General.AuthController.VerifyLogin)
	}

	// Public invitation routes (no authentication required)
	invitations := routerGroup.Group("/invitations")
	{
		invitations.GET("", app.Controllers.Customer.TeamController.GetPendingInvitations)
		invitations.POST("/accept-team", app.Controllers.Customer.TeamController.AcceptTeamInvitation)
		invitations.POST("/reject-team", app.Controllers.Customer.TeamController.RejectTeamInvitation)
	}
}
