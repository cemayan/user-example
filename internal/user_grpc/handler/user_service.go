package handler

import (
	"encoding/json"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/dto"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/model"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/repo"
	pb "github.com/cemayan/faceit-technical-test/protos/event"
	"github.com/sirupsen/logrus"
)

type UserEventHandler interface {
	Handle() error
}

type UserHandler struct {
	userRepo    repo.GrpcUserRepository
	event       *pb.Events
	eventServer pb.EventGrpcService_HandleEventServer
	log         *logrus.Logger
}

// Handle consumes the stream events
// When request come to CREATE endpoint of API  these events called
func (uh UserHandler) Handle() error {

	switch uh.event.EventName {
	case pb.EventName_USER_CREATED:

		var user model.User
		err := json.Unmarshal(uh.event.EventData, &user)
		if err != nil {
			uh.log.WithFields(logrus.Fields{"method": "Handle"}).Errorln("An error occurred when unmarshalling the incoming eventdata")
			return err
		}

		createUser, err := uh.userRepo.CreateUser(&user)
		if err != nil {
			uh.eventServer.Send(&pb.Response{
				Data:       []byte(err.Error()),
				StatusCode: 400,
			})

		} else {
			response, _ := json.Marshal(createUser)
			err := uh.eventServer.Send(&pb.Response{
				Data:       response,
				StatusCode: 200,
			})
			if err != nil {
				return err
			}
		}

	case pb.EventName_USER_UPDATED:
		var user dto.UpdateUser
		err := json.Unmarshal(uh.event.EventData, &user)
		if err != nil {
			uh.log.WithFields(logrus.Fields{"method": "Handle-Create"}).Errorln("An error occurred when unmarshalling the incoming eventdata")
			return err
		}

		err = uh.userRepo.UpdateUser(uh.event.InternalId, &user)
		if err != nil {
			return err
		} else {
			err := uh.eventServer.Send(&pb.Response{
				StatusCode: 200,
			})
			if err != nil {
				uh.log.WithFields(logrus.Fields{"method": "Handle-Update"}).Errorln("An error occurred when unmarshalling the incoming eventdata")
				return err
			}
		}
	case pb.EventName_USER_DELETED:
		err := uh.userRepo.DeleteUser(uh.event.InternalId)
		if err != nil {
			return err
		} else {
			err := uh.eventServer.Send(&pb.Response{
				StatusCode: 200,
			})
			if err != nil {
				uh.log.WithFields(logrus.Fields{"method": "Handle-Delete"}).Errorln("An error occurred when unmarshalling the incoming eventdata")
				return err
			}
		}
	}
	return nil
}

func NewUserEventHandler(userRepo repo.GrpcUserRepository, event *pb.Events, eventServer pb.EventGrpcService_HandleEventServer, log *logrus.Logger) UserEventHandler {
	return &UserHandler{
		userRepo:    userRepo,
		event:       event,
		eventServer: eventServer,
		log:         log,
	}
}
