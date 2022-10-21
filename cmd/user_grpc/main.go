package main

import (
	"fmt"
	"github.com/cemayan/faceit-technical-test/config/user"
	_ "github.com/cemayan/faceit-technical-test/docs/user_grpc"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/database"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/router"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/util"
	"github.com/cemayan/faceit-technical-test/pkg/postgres"
	pb "github.com/cemayan/faceit-technical-test/protos/event"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
)

var _log *logrus.Logger
var app = fiber.New()
var configs *user.AppConfig
var v *viper.Viper
var dbHandler postgres.DBHandler
var grpcConn *grpc.ClientConn

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
		_log.WithFields(logrus.Fields{"service": "user_grpc"}).Errorf("An error occured when getting config. %v", err)
		return
	}

	str := fmt.Sprintf("%s:%s", configs.Grpc.ADDR, configs.Grpc.PORT)
	//gRPC connection
	_grpcConn, err := grpc.Dial(str, grpc.WithTransportCredentials(insecure.NewCredentials()))
	grpcConn = _grpcConn
	util.FailOnError(err, "did not connect")

	_log.Infoln("gRPC connection is starting...")

	//Postresql connection
	dbHandler = postgres.NewDbHandler(&configs.Postgresql, _log.WithFields(logrus.Fields{"service": "user_grpc"}))
	_db := dbHandler.New()
	database.DB = _db
	util.MigrateDB(_db, _log.WithFields(logrus.Fields{"service": "user_grpc"}))
}

// @title        Faceit
// @version      1.0
// @description  This is a swagger for Faceit
// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html
// @host         localhost:8092
// @BasePath     /api/v1/user
// @securityDefinitions.apikey Bearer
// @in                         header
// @Tags                       User
// @name                       Authorization
func main() {

	grpcClient := pb.NewEventGrpcServiceClient(grpcConn)

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

	router.SetupGrpcRoutes(app, _log.WithFields(logrus.Fields{"service": "user_grpc"}), grpcClient, configs)

	err := app.Listen(":8092")
	if err != nil {
		return
	}

}
