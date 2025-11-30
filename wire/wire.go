//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/ELMOCPC/Backend/bootstrap"
	"github.com/ELMOCPC/Backend/internal/application/service"
	"github.com/ELMOCPC/Backend/internal/application/usecase"
	domainCache "github.com/ELMOCPC/Backend/internal/domain/repository/cache"
	domainPostgre "github.com/ELMOCPC/Backend/internal/domain/repository/postgres"
	domainS3 "github.com/ELMOCPC/Backend/internal/domain/storage/s3"
	infraCache "github.com/ELMOCPC/Backend/internal/infrastructure/cache"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
	infraJWT "github.com/ELMOCPC/Backend/internal/infrastructure/jwt"
	infraLocalization "github.com/ELMOCPC/Backend/internal/infrastructure/localization"
	infraPostgre "github.com/ELMOCPC/Backend/internal/infrastructure/repository/postgres"
	seed "github.com/ELMOCPC/Backend/internal/infrastructure/seed"
	infraStorage "github.com/ELMOCPC/Backend/internal/infrastructure/storage"
	"github.com/ELMOCPC/Backend/internal/presentation/controller"
	"github.com/ELMOCPC/Backend/internal/presentation/middleware"
	"github.com/google/wire"
)

var DatabaseProviderSet = wire.NewSet(
	database.NewPostgresDatabase,
	database.NewRedisDatabase,
	wire.Bind(new(database.Database), new(*database.PostgresDatabase)),
	wire.Bind(new(database.Cache), new(*database.RedisDatabase)),
	wire.Struct(new(Database), "*"),
)

var CacheProviderSet = wire.NewSet(
	bootstrap.ProvideRedisClient,
	infraCache.NewRedisCache,
	wire.Bind(new(domainCache.CacheRepository), new(*infraCache.RedisCache)),
)

var RepositoryProviderSet = wire.NewSet(
	infraPostgre.NewUserRepository,
	wire.Bind(new(domainPostgre.UserRepository), new(*infraPostgre.UserRepository)),
	infraPostgre.NewTeamRepository,
	wire.Bind(new(domainPostgre.TeamRepository), new(*infraPostgre.TeamRepository)),
	infraPostgre.NewOTPRepository,
	infraPostgre.NewTeamInvitationTokenRepository,
	infraPostgre.NewPaymentReceiptRepository,
)

var ServiceProviderSet = wire.NewSet(
	service.NewJWTService,
	wire.Bind(new(usecase.JwtService), new(*service.JWTService)),
	service.NewEmailService,
	wire.Bind(new(usecase.EmailService), new(*service.EmailService)),
	service.NewOTPService,
	service.NewAuthService,
	wire.Bind(new(usecase.AuthService), new(*service.AuthService)),
	service.NewTeamService,
	wire.Bind(new(usecase.TeamService), new(*service.TeamService)),
	service.NewTeamInvitationService,
	service.NewPaymentReceiptService,
)

var GeneralControllerProviderSet = wire.NewSet(
	controller.NewAuthController,
	wire.Struct(new(GeneralControllers), "*"),
)

var AdminControllerProviderSet = wire.NewSet(
	wire.Struct(new(AdminControllers), "*"),
)

var CustomerControllerProviderSet = wire.NewSet(
	controller.NewTeamController,
	wire.Struct(new(CustomerControllers), "*"),
)

var ControllerProviderSet = wire.NewSet(
	wire.Struct(new(Controllers), "*"),
)

var AdapterProviderSet = wire.NewSet(
	infraJWT.NewJWTKeyManager,
	infraLocalization.NewTranslationService,
	infraStorage.NewS3Storage,
	wire.Bind(new(domainS3.S3Storage), new(*infraStorage.S3Storage)),
)

var MiddlewareProviderSet = wire.NewSet(
	middleware.NewCorsMiddleware,
	middleware.NewRecoveryMiddleware,
	middleware.NewLocalizationMiddleware,
	middleware.NewAuthMiddleware,
	middleware.NewRateLimitMiddleware,
	wire.Struct(new(Middlewares), "*"),
)

var SeedProviderSet = wire.NewSet(
	seed.NewRoleSeeder,
	wire.Struct(new(Seeds), "*"),
)

func ProvideDBConfig(container *bootstrap.Config) *bootstrap.Database {
	return &container.Env.Database
}

func ProvideConstants(container *bootstrap.Config) *bootstrap.Constants {
	return container.Constants
}

func ProvideRDBConfig(container *bootstrap.Config) *bootstrap.Redis {
	return &container.Env.Cache
}

func ProvideOTPConfig(container *bootstrap.Config) *bootstrap.OTP {
	return &container.Env.OTP
}

func ProvideStorageConfig(container *bootstrap.Config) *bootstrap.S3 {
	return &container.Env.S3
}

func ProvideJWTKeysPath(container *bootstrap.Config) *bootstrap.JWTKeysPath {
	return &container.Constants.JWTKeysPath
}

func ProvideSuperAdminCredentials(container *bootstrap.Config) *bootstrap.SuperAdmin {
	return &container.Env.SuperAdmin
}

func ProvidePagination(container *bootstrap.Config) *bootstrap.Pagination {
	return &container.Env.Pagination
}

func ProvideSMTPConfig(container *bootstrap.Config) *service.EmailConfig {
	return &service.EmailConfig{
		SMTPServer:     container.Env.SMTP.Server,
		SMTPPort:       container.Env.SMTP.Port,
		SenderEmail:    container.Env.SMTP.Email,
		SenderPassword: container.Env.SMTP.Password,
		SenderName:     container.Env.SMTP.SenderName,
	}
}

func ProvideEnv(container *bootstrap.Config) *bootstrap.Env {
	return container.Env
}

var ProviderSet = wire.NewSet(
	DatabaseProviderSet,
	CacheProviderSet,
	RepositoryProviderSet,
	ServiceProviderSet,
	MiddlewareProviderSet,
	GeneralControllerProviderSet,
	AdminControllerProviderSet,
	CustomerControllerProviderSet,
	ControllerProviderSet,
	AdapterProviderSet,
	ProvideEnv,
	ProvideDBConfig,
	ProvideConstants,
	ProvideRDBConfig,
	ProvideOTPConfig,
	ProvideStorageConfig,
	ProvideJWTKeysPath,
	ProvideSuperAdminCredentials,
	ProvidePagination,
	ProvideSMTPConfig,
	SeedProviderSet,
)

type Database struct {
	DB  database.Database
	RDB database.Cache
}

type GeneralControllers struct {
	AuthController *controller.AuthController
}

type AdminControllers struct {
}

type CustomerControllers struct {
	TeamController *controller.TeamController
}

type Controllers struct {
	General  *GeneralControllers
	Admin    *AdminControllers
	Customer *CustomerControllers
}

type Middlewares struct {
	Cors         *middleware.CORSMiddleware
	Recovery     *middleware.RecoveryMiddleware
	Localization *middleware.LocalizationMiddleware
	Auth         *middleware.AuthMiddleware
	RateLimit    *middleware.RateLimitMiddleware
}

type Seeds struct {
	RoleSeeder *seed.RoleSeeder
}

type Application struct {
	Database    *Database
	Middlewares *Middlewares
	Controllers *Controllers
	Seeds       *Seeds
}

func NewApplication(
	database *Database,
	middlewares *Middlewares,
	controllers *Controllers,
	seeds *Seeds,
) *Application {
	return &Application{
		Database:    database,
		Middlewares: middlewares,
		Controllers: controllers,
		Seeds:       seeds,
	}
}

func InitializeApplication(config *bootstrap.Config) (*Application, error) {
	wire.Build(
		ProviderSet,
		NewApplication,
	)
	return &Application{}, nil
}
