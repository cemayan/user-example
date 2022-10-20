package router

import (
	"github.com/cemayan/faceit-technical-test/config/user"
	"github.com/cemayan/faceit-technical-test/internal/user/database"
	"github.com/cemayan/faceit-technical-test/internal/user/repo"
	"github.com/cemayan/faceit-technical-test/internal/user/service"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func prometheusHandler2() http.Handler {
	return promhttp.Handler()
}

// SetupAuthRoutes creates the fiber's routes
// api/v1 is root group.
// Before the reach services interface is configured
func SetupAuthRoutes(app *fiber.App, log *log.Logger, configs *user.AppConfig) {

	api := app.Group("/api", logger.New())
	v1 := api.Group("/v1")

	v1.Get("/metrics", adaptor.HTTPHandler(prometheusHandler2()))

	v1.Get("/swagger/*", swagger.HandlerDefault) // default

	userRepo := repo.NewUserRepo(database.DB, log)

	var authSvc service.AuthService
	authSvc = service.NewAuthService(userRepo, log, configs)

	v1.Get("/health", authSvc.HealthCheck)

	//Auth
	auth := v1.Group("/auth")
	auth.Post("/getToken", authSvc.Login)

}
