package main

import (
	"github.com/cemayan/faceit-technical-test/config/user"
	_ "github.com/cemayan/faceit-technical-test/docs/auth"
	"github.com/cemayan/faceit-technical-test/internal/user/database"
	"github.com/cemayan/faceit-technical-test/internal/user/router"
	"github.com/cemayan/faceit-technical-test/internal/user/util"
	"github.com/cemayan/faceit-technical-test/pkg/postgres"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

var _log *logrus.Logger
var app = fiber.New()
var configs *user.AppConfig
var v *viper.Viper
var dbHandler postgres.DBHandler

func init() {
	//logrus init
	_log = logrus.New()
	_log.Out = os.Stdout

	v = viper.New()
	_configs := user.NewConfig(v)

	env := os.Getenv("ENV")
	appConfig, err := _configs.GetConfig(env)
	configs = appConfig
	if err != nil {
		return
	}

	//Postresql connection
	dbHandler = postgres.NewDbHandler(&configs.Postgresql)
	_db := dbHandler.New()
	database.DB = _db
	util.MigrateDB(_db)
}

// @title        Faceit
// @version      1.0
// @description  This is a swagger for Faceit
// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html
// @host         localhost:8109
// @BasePath     /api/v1/auth
func main() {

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	_log.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	router.SetupAuthRoutes(app, _log, configs)

	err := app.Listen(":8109")
	if err != nil {
		return
	}

}
