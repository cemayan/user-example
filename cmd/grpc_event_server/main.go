package main

import (
	"fmt"
	"github.com/cemayan/faceit-technical-test/config/user"
	"github.com/cemayan/faceit-technical-test/internal/grpc_event_server/handler"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/database"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/repo"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/util"
	"github.com/cemayan/faceit-technical-test/pkg/postgres"
	pb "github.com/cemayan/faceit-technical-test/protos/event"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"io"
	"net"
	"os"
)

type server struct {
	pb.UnimplementedEventGrpcServiceServer
}

var configs *user.AppConfig
var v *viper.Viper
var _log *logrus.Logger
var grpcConn *grpc.ClientConn
var dbHandler postgres.DBHandler
var userRepo repo.GrpcUserRepository
var userEventHandler handler.UserEventHandler

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
		_log.WithFields(logrus.Fields{"service": "grpc_event_server"}).Errorf("An error occured when getting config. %v", err)

		return
	}

	//Postresql connection
	dbHandler = postgres.NewDbHandler(&configs.Postgresql, _log.WithFields(logrus.Fields{"service": "grpc_event_server"}))
	_db := dbHandler.New()
	database.DB = _db
	util.MigrateDB(_db, _log.WithFields(logrus.Fields{"service": "grpc_event_server"}))
	userRepo = repo.NewGrpcUserRepo(database.DB, _log.WithFields(logrus.Fields{"service": "grpc_event_server"}))
}

func (s server) HandleEvent(eventServer pb.EventGrpcService_HandleEventServer) error {
	_log.WithFields(logrus.Fields{"service": "grpc_event_server"}).Infoln("Event consuming operation is starting ...")
	for {
		event, err := eventServer.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			_log.WithFields(logrus.Fields{"service": "grpc_event_server"}).Errorf("An error occured %s", err.Error())
		}

		switch event.AggregateType {
		case pb.AggregateType_USER:
			userEventHandler = handler.NewUserEventHandler(userRepo, event, eventServer, _log)
			err = userEventHandler.Handle()
			if err != nil {
				_log.WithFields(logrus.Fields{"service": "grpc_event_server"}).Errorf("An error occured %s", err.Error())
			}
		}
	}

	return nil
}

func main() {

	if os.Getenv("ENV") == "dev" {

		_log.SetFormatter(&logrus.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})
	} else {
		_log.SetFormatter(&logrus.JSONFormatter{})
		_log.SetOutput(os.Stdout)
		_log.WithFields(logrus.Fields{"service": "grpc_event_server"})
	}

	_log.Infoln("gRPC server is starting...")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", configs.Grpc.PORT))
	util.FailOnError(err, "tcp listen failed.")

	// gRPC implementation
	s := grpc.NewServer()
	pb.RegisterEventGrpcServiceServer(s, &server{})
	_log.WithFields(logrus.Fields{"service": "grpc_event_server"}).Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		_log.WithFields(logrus.Fields{"service": "grpc_event_server"}).Errorf("failed to serve: %v", err)
	}

}
