package httpv1

import (
	"github.com/ELMOCPC/Backend/wire"
	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(routerGroup *gin.RouterGroup, app *wire.Application) {

	teams := routerGroup.Group("/teams")
	{
		teams.GET("", app.Controllers.Customer.TeamController.GetAllTeams)
		teams.GET("/:id", app.Controllers.Customer.TeamController.GetTeamByIDAdmin)
		teams.POST("/:id/validate", app.Controllers.Customer.TeamController.ValidateTeam)
		teams.POST("/:id/reject", app.Controllers.Customer.TeamController.RejectTeam)
	}

	receipts := routerGroup.Group("/receipts")
	{
		receipts.GET("/pending", app.Controllers.Customer.TeamController.GetPendingReceipts)
		receipts.POST("/:id/verify", app.Controllers.Customer.TeamController.VerifyPaymentReceipt)
	}

	users := routerGroup.Group("/users")
	{
		users.GET("", app.Controllers.General.AuthController.GetAllUsers)
	}

}
