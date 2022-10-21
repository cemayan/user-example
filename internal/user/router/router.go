package router

import (
	"github.com/cemayan/faceit-technical-test/config/user"
	"github.com/cemayan/faceit-technical-test/internal/user/database"
	"github.com/cemayan/faceit-technical-test/internal/user/repo"
	"github.com/cemayan/faceit-technical-test/internal/user/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func prometheusHandler() http.Handler {
	return promhttp.Handler()
}

// SetupRoutes creates the fiber's routes
// api/v1 is root group.
// Before the reach services interface is configured
func SetupRoutes(app *fiber.App, log *log.Entry, configs *user.AppConfig) {

	app.Get("/metrics", monitor.New(monitor.Config{Title: "MyService Metrics Page"}))

	api := app.Group("/api", logger.New())
	v1 := api.Group("/v1")

	v1.Get("/metrics", adaptor.HTTPHandler(prometheusHandler()))
	v1.Get("/swagger/*", swagger.HandlerDefault) // default

	userRepo := repo.NewUserRepo(database.DB, log)

	var validate *validator.Validate
	validate = validator.New()
	var userSvc service.UserService
	userSvc = service.NewUserService(userRepo, validate, log, configs)

	v1.Get("/health", userSvc.HealthCheck)

	userGroup := v1.Group("/user")
	userGroup.Get("/", userSvc.GetAllUser)
	userGroup.Get("/:id", userSvc.GetUser)
	userGroup.Post("/", userSvc.CreateUser)
	userGroup.Put("/:id", userSvc.UpdateUser)
	userGroup.Delete("/:id", userSvc.DeleteUser)
}
