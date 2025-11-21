package httpv1

import (
	"github.com/ELMOCPC/Backend/wire"
	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(routerGroup *gin.RouterGroup, app *wire.Application) {
	// Payment receipt routes
	receipts := routerGroup.Group("/receipts")
	{
		receipts.GET("/pending", app.Controllers.Customer.TeamController.GetPendingReceipts)
		receipts.POST("/:id/verify", app.Controllers.Customer.TeamController.VerifyPaymentReceipt)
	}
}
