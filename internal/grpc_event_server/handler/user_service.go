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

func (uh UserHandler) Handle() error {

	switch uh.event.EventName {
	case pb.EventName_USER_CREATED:

		var user model.User
		err := json.Unmarshal(uh.event.EventData, &user)
		if err != nil {
			uh.log.Errorln("An error occurred when unmarshalling the incoming eventdata")
			return err
		}

		createUser, err := uh.userRepo.CreateUser(&user)
		if err != nil {
			return err
		} else {

			response, _ := json.Marshal(createUser)
			uh.eventServer.Send(&pb.Response{
				Data:       response,
				StatusCode: 200,
			})
		}

	case pb.EventName_USER_UPDATED:
		var user dto.UpdateUser
		err := json.Unmarshal(uh.event.EventData, &user)
		if err != nil {
			uh.log.Errorln("An error occurred when unmarshalling the incoming eventdata")
			return err
		}

		err = uh.userRepo.UpdateUser(uh.event.InternalId, &user)
		if err != nil {
			return err
		} else {
			uh.eventServer.Send(&pb.Response{
				StatusCode: 200,
			})
		}
	case pb.EventName_USER_DELETED:
		err := uh.userRepo.DeleteUser(uh.event.InternalId)
		if err != nil {
			return err
		} else {
			uh.eventServer.Send(&pb.Response{
				StatusCode: 200,
			})
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