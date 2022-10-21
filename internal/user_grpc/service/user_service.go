package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cemayan/faceit-technical-test/config/user"
	_ "github.com/cemayan/faceit-technical-test/internal/user_grpc/dto"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/model"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/repo"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/util"
	"github.com/cemayan/faceit-technical-test/pkg/common"
	"github.com/cemayan/faceit-technical-test/protos/event"
	pb "github.com/cemayan/faceit-technical-test/protos/event"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type GrpcUserService interface {
	GetUser(c *fiber.Ctx) error
	GetAllUser(c *fiber.Ctx) error
	CreateUser(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
}

// A UserSvc  contains the required dependencies for this service
type GrpcUserSvc struct {
	repository repo.GrpcUserRepository
	log        *log.Logger
	grpcClient pb.EventGrpcServiceClient
	configs    *user.AppConfig
}

// GetAllUser returns filtered users based on given payload
// @Summary  GetAllUser
// @Param    limit path number false "limit"
// @Param    page path number false "page"
// @Param    sColumn path number false "sColumn"
// @Param    sType path number false "sType"
// @Param    cQuery path string false "cQuery"
// @Param    cVal path string false "cVal"
// @Tags     User
// @Router   / [get]
// @Security Bearer
func (s GrpcUserSvc) GetAllUser(c *fiber.Ctx) error {
	var pagination common.Pagination
	err := c.QueryParser(&pagination)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message:    fmt.Sprintf("An error occured %s", err),
			StatusCode: 400,
		})
	}

	result, err := s.repository.GetAllUser(pagination)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message:    fmt.Sprintf("An error occured %s", err),
			StatusCode: 400,
		})
	}

	return c.JSON(&model.Response{
		Data:       result,
		StatusCode: 200,
	})
}

// HealthCheck returns 200 with body
func (s GrpcUserSvc) HealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("healty!")
}

// GetUser returns user based on given id.
// @Summary  GetUser
// @Param    id path string true "id"
// @Tags     User
// @Router   /{id} [get]
// @Security Bearer
func (s GrpcUserSvc) GetUser(c *fiber.Ctx) error {

	id := c.Params("id")
	user, err := s.repository.GetUserById(id)
	if user == nil || err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message:    fmt.Sprintf("No user found with %v", id),
			StatusCode: 400,
		})
	}

	return c.JSON(model.Response{
		StatusCode: 200,
		Data: model.UserData{
			ID:        user.ID,
			NickName:  user.NickName,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Country:   user.Country,
		},
	})
}

// CreateUser creates new user based on given payload
// While user is creating password is encrypted then it is assigned as a password
// @Summary  CreateUser
// @Param    request body model.User true "query params"
// @Tags     User
// @Router   / [post]
func (s GrpcUserSvc) CreateUser(c *fiber.Ctx) error {

	handleEvent, err := s.grpcClient.HandleEvent(context.Background())
	if err != nil {
		return err
	}

	event := &pb.Events{
		AggregateId:   uuid.New().String(),
		AggregateType: 0,
		EventData:     c.Body(),
		EventDate:     util.GetTime(),
		EventName:     pb.EventName_USER_CREATED,
	}
	handleEvent.Send(event)

	recv, err := handleEvent.Recv()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(&model.Response{
			Message:    err.Error(),
			StatusCode: 400,
		})
	} else {
		var user model.User
		json.Unmarshal(recv.Data, &user)
		return c.Status(fiber.StatusCreated).JSON(&model.Response{
			Message: "User created!",
			Data: model.UserData{
				ID:        user.ID,
				NickName:  user.NickName,
				Email:     user.Email,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Country:   user.Country,
			},
			StatusCode: 201,
		})
	}
}

// UpdateUser return updated user based on given payload
// @Summary  UpdateUser
// @Param    id      path string         true "id"
// @Param    request body dto.UpdateUser true "query params"
// @Tags     User
// @Router   /{id} [put]
// @Security Bearer
func (s GrpcUserSvc) UpdateUser(c *fiber.Ctx) error {

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			StatusCode: 400,
			Message:    fmt.Sprintf("Review your id %v", id),
		})
	}

	handleEvent, err := s.grpcClient.HandleEvent(context.Background())
	if err != nil {
		return err
	}

	event := &pb.Events{
		AggregateId:   uuid.New().String(),
		AggregateType: 0,
		EventData:     c.Body(),
		EventDate:     util.GetTime(),
		EventName:     pb.EventName_USER_UPDATED,
		InternalId:    id,
	}
	handleEvent.Send(event)

	_, err = handleEvent.Recv()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(&model.Response{
			Message:    err.Error(),
			StatusCode: 400,
		})
	} else {

		return c.Status(fiber.StatusOK).JSON(&model.Response{
			Message:    "User updated!",
			StatusCode: 200,
		})
	}

}

// DeleteUser removes  the user based on given payload
// @Summary  DeleteUser
// @Param    id path string true "id"
// @Tags     User
// @Router   /{id} [delete]
// @Security Bearer
func (s GrpcUserSvc) DeleteUser(c *fiber.Ctx) error {

	id := c.Params("id")

	handleEvent, err := s.grpcClient.HandleEvent(context.Background())
	if err != nil {
		return err
	}

	event := &pb.Events{
		AggregateId:   uuid.New().String(),
		AggregateType: 0,
		EventData:     c.Body(),
		EventDate:     util.GetTime(),
		EventName:     pb.EventName_USER_DELETED,
		InternalId:    id,
	}
	handleEvent.Send(event)

	_, err = handleEvent.Recv()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(&model.Response{
			Message:    err.Error(),
			StatusCode: 400,
		})
	} else {

		return c.Status(fiber.StatusOK).JSON(&model.Response{
			Message:    "User successfully deleted!",
			StatusCode: 200,
		})
	}

}

func NewGrpcUserService(rep repo.GrpcUserRepository, grpcClient event.EventGrpcServiceClient, log *log.Logger, configs *user.AppConfig) GrpcUserService {
	return &GrpcUserSvc{
		repository: rep,
		grpcClient: grpcClient,
		log:        log,
		configs:    configs,
	}
}
