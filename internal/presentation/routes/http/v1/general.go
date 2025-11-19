package httpv1

import (
	"github.com/ELMOCPC/Backend/wire"
	"github.com/gin-gonic/gin"
)

func SetupGeneralRoutes(routerGroup *gin.RouterGroup, app *wire.Application) {
	auth := routerGroup.Group("/auth")
	{
		auth.POST("/signup", app.Controllers.General.AuthController.SignUp)
		auth.POST("/login", app.Controllers.General.AuthController.Login)
	}
}
