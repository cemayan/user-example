package main

import (
	"github.com/cemayan/faceit-technical-test/config/user"
	_ "github.com/cemayan/faceit-technical-test/docs/user"
	"github.com/cemayan/faceit-technical-test/internal/user/database"
	"github.com/cemayan/faceit-technical-test/internal/user/router"
	"github.com/cemayan/faceit-technical-test/internal/user/util"
	"github.com/cemayan/faceit-technical-test/pkg/postgres"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
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
		_log.WithFields(logrus.Fields{"service": "user"}).Errorf("An error occured when getting config. %v", err)

		return
	}

	//Postresql connection
	dbHandler = postgres.NewDBHandler(&configs.Postgresql, _log.WithFields(logrus.Fields{"service": "user"}))
	_db := dbHandler.New()
	database.DB = _db
	util.MigrateDB(_db, _log.WithFields(logrus.Fields{"service": "user"}))
}

// @title        Faceit
// @version      1.0
// @description  This is a swagger for Faceit
// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html
// @host         localhost:8089
// @BasePath     /api/v1/user
func main() {

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	if os.Getenv("ENV") == "dev" {
		_log.SetFormatter(&logrus.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})
	} else {
		_log.SetFormatter(&logrus.JSONFormatter{})
		_log.SetOutput(os.Stdout)
	}

	_log.AddHook(&writer.Hook{ // Send info and debug logs to stdout
		Writer: os.Stdout,
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
		},
	})
	router.SetupRoutes(app, _log.WithFields(logrus.Fields{"service": "user"}), configs)

	err := app.Listen(":8089")
	if err != nil {
		return
	}

}
