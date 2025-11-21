package httpv1

import (
	"github.com/ELMOCPC/Backend/wire"
	"github.com/gin-gonic/gin"
)

func SetupCustomerRoutes(routerGroup *gin.RouterGroup, app *wire.Application) {
	// Team routes
	teams := routerGroup.Group("/teams")
	{
		teams.POST("", app.Controllers.Customer.TeamController.CreateTeam)
		teams.GET("", app.Controllers.Customer.TeamController.GetUserTeam)
		teams.GET("/:id", app.Controllers.Customer.TeamController.GetTeam)
		teams.PUT("/:id", app.Controllers.Customer.TeamController.UpdateTeam)
		teams.DELETE("/:id", app.Controllers.Customer.TeamController.DeleteTeam)
		teams.GET("/:id/members", app.Controllers.Customer.TeamController.GetTeamMembers)
		teams.POST("/:id/leave", app.Controllers.Customer.TeamController.LeaveTeam)

		teams.POST("/:id/invitations", app.Controllers.Customer.TeamController.InviteMembers)
		teams.GET("/:id/invitations/pending", app.Controllers.Customer.TeamController.GetPendingTeamInvitations)
		teams.DELETE("/:id/invitations/:tokenId", app.Controllers.Customer.TeamController.CancelTeamInvitation)
		teams.DELETE("/:id/members/:memberId", app.Controllers.Customer.TeamController.RemoveMember)

		teams.POST("/:id/upload-receipt", app.Controllers.Customer.TeamController.UploadPaymentReceipt)
		teams.GET("/:id/receipt", app.Controllers.Customer.TeamController.GetTeamReceipt)
		teams.POST("/:id/submit", app.Controllers.Customer.TeamController.SubmitTeam)
	}
}
