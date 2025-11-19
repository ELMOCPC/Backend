package httpv1

import (
	"github.com/ELMOCPC/Backend/wire"
	"github.com/gin-gonic/gin"
)

func SetupCustomerRoutes(routerGroup *gin.RouterGroup, app *wire.Application) {
	// Add your customer routes here
	// Example:
	// profile := routerGroup.Group("/profile")
	// {
	//     profile.GET("", app.Controllers.Customer.UserController.GetProfile)
	//     profile.PUT("", app.Controllers.Customer.UserController.UpdateProfile)
	// }
}
