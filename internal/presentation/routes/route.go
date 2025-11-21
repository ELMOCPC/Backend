package routes

import (
	httpv1 "github.com/ELMOCPC/Backend/internal/presentation/routes/http/v1"
	"github.com/ELMOCPC/Backend/wire"
	"github.com/gin-gonic/gin"
)

func Run(ginEngine *gin.Engine, app *wire.Application) {
	ginEngine.Use(gin.Logger())
	ginEngine.Use(app.Middlewares.Cors.CORS())
	ginEngine.Use(app.Middlewares.Recovery.Recovery)
	ginEngine.Use(app.Middlewares.Localization.Localization)

	v1 := ginEngine.Group("/v1")
	registerGeneralRoutes(v1, app)
	registerCustomerRoutes(v1, app)
	registerAdminRoutes(v1, app)
}

func registerGeneralRoutes(v1 *gin.RouterGroup, app *wire.Application) {
	httpv1.SetupGeneralRoutes(v1, app)
}

func registerCustomerRoutes(v1 *gin.RouterGroup, app *wire.Application) {
	user := v1.Group("/user")
	user.Use(app.Middlewares.Auth.AuthRequired)
	httpv1.SetupCustomerRoutes(user, app)
}

func registerAdminRoutes(v1 *gin.RouterGroup, app *wire.Application) {
	admin := v1.Group("/admin")
	admin.Use(app.Middlewares.Auth.AuthRequired)
	httpv1.SetupAdminRoutes(admin, app)
}
