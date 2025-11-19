package httpv1

import (
	"github.com/ELMOCPC/Backend/wire"
	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(routerGroup *gin.RouterGroup, app *wire.Application) {
	// Add your admin routes here
	// Example:
	// users := routerGroup.Group("/users")
	// {
	//     users.GET("", app.Controllers.Admin.UserController.GetUsers)
	//     users.POST("", app.Controllers.Admin.UserController.CreateUser)
	// }
}
