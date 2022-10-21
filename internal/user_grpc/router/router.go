package router

import (
	"github.com/cemayan/faceit-technical-test/config/user"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/database"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/repo"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/service"
	pb "github.com/cemayan/faceit-technical-test/protos/event"
	"github.com/go-playground/validator/v10"
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
func SetupGrpcRoutes(app *fiber.App, _log *log.Entry, client pb.EventGrpcServiceClient, configs *user.AppConfig) {

	api := app.Group("/api", logger.New())
	v1 := api.Group("/v1")

	v1.Get("/metrics", adaptor.HTTPHandler(grpcPrometheusHandler()))

	v1.Get("/swagger/*", swagger.HandlerDefault) // default

	userRepo := repo.NewGrpcUserRepo(database.DB, _log)

	var validate *validator.Validate
	validate = validator.New()

	var userSvc service.GrpcUserService
	userSvc = service.NewGrpcUserService(userRepo, validate, client, _log, configs)

	userGroup := v1.Group("/user")
	userGroup.Get("/", userSvc.GetAllUser)
	userGroup.Get("/:id", userSvc.GetUser)
	userGroup.Post("/", userSvc.CreateUser)
	userGroup.Put("/:id", userSvc.UpdateUser)
	userGroup.Delete("/:id", userSvc.DeleteUser)
}
