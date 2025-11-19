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
		teams.POST("/:id/submit", app.Controllers.Customer.TeamController.SubmitTeam)
		teams.GET("/:id/members", app.Controllers.Customer.TeamController.GetTeamMembers)
		teams.POST("/:id/invitations", app.Controllers.Customer.TeamController.InviteUser)
		teams.DELETE("/:id/members/:memberId", app.Controllers.Customer.TeamController.RemoveMember)
		teams.POST("/:id/leave", app.Controllers.Customer.TeamController.LeaveTeam)
	}

	// Invitation routes
	invitations := routerGroup.Group("/invitations")
	{
		invitations.GET("", app.Controllers.Customer.TeamController.GetUserInvitations)
		invitations.POST("/:id/accept", app.Controllers.Customer.TeamController.AcceptInvitation)
		invitations.POST("/:id/reject", app.Controllers.Customer.TeamController.RejectInvitation)
		invitations.DELETE("/:id", app.Controllers.Customer.TeamController.CancelInvitation)
	}
}
