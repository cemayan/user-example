package router

import (
	"github.com/cemayan/faceit-technical-test/config/user"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/database"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/middleware"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/repo"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/service"
	pb "github.com/cemayan/faceit-technical-test/protos/event"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func grpcPrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// SetupRoutes creates the fiber's routes
// api/v1 is root group.
// Before the reach services interface is configured
func SetupGrpcRoutes(app *fiber.App, log *log.Logger, client pb.EventGrpcServiceClient, configs *user.AppConfig) {

	api := app.Group("/api", logger.New())
	v1 := api.Group("/v1")

	v1.Get("/metrics", adaptor.HTTPHandler(grpcPrometheusHandler()))

	v1.Get("/swagger/*", swagger.HandlerDefault) // default

	userRepo := repo.NewGrpcUserRepo(database.DB, log)

	var authSvc service.GrpcAuthService
	authSvc = service.NewAuthService(userRepo, log, configs)

	var userSvc service.GrpcUserService
	userSvc = service.NewGrpcUserService(userRepo, authSvc, client, log, configs)

	userGroup := v1.Group("/user")
	userGroup.Get("/", middleware.Protected(configs), userSvc.GetAllUser)
	userGroup.Get("/:id?", middleware.Protected(configs), userSvc.GetUser)
	userGroup.Post("/", userSvc.CreateUser)
	userGroup.Put("/:id", middleware.Protected(configs), userSvc.UpdateUser)
	userGroup.Delete("/:id", middleware.Protected(configs), userSvc.DeleteUser)
}
