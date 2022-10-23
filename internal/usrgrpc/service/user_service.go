package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cemayan/faceit-technical-test/config/user"
	_ "github.com/cemayan/faceit-technical-test/internal/usrgrpc/dto"
	"github.com/cemayan/faceit-technical-test/internal/usrgrpc/model"
	"github.com/cemayan/faceit-technical-test/internal/usrgrpc/repo"
	"github.com/cemayan/faceit-technical-test/internal/usrgrpc/util"
	"github.com/cemayan/faceit-technical-test/pkg/common"
	pb "github.com/cemayan/faceit-technical-test/protos/event"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type GrpcUserService interface {
	HashPassword(password string) (string, error)
	GetUser(c *fiber.Ctx) error
	GetAllUser(c *fiber.Ctx) error
	CreateUser(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
}

// A GrpcUserSvc  contains the required dependencies for this service
type GrpcUserSvc struct {
	repository repo.GrpcUserRepository
	validate   *validator.Validate
	log        *log.Entry
	grpcClient pb.EventGrpcServiceClient
	configs    *user.AppConfig
}

// HashPassword returns encrypted password based on given password
func (s GrpcUserSvc) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
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
func (s GrpcUserSvc) GetAllUser(c *fiber.Ctx) error {
	var pagination common.Pagination
	err := c.QueryParser(&pagination)
	if err != nil {
		s.log.WithFields(log.Fields{"method": "GetAllUser"}).Errorf(fmt.Sprintf("An error occured %s \n", err))
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message:    fmt.Sprintf("An error occured %s", err),
			StatusCode: 400,
		})
	}

	result, err := s.repository.GetAllUser(pagination)

	if err != nil {
		s.log.WithFields(log.Fields{"method": "GetAllUser"}).Errorf(fmt.Sprintf("An error occured %s \n", err))
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
	return c.Status(fiber.StatusOK).JSON("UP!")
}

// GetUser returns user based on given id.
// @Summary  GetUser
// @Param    id path string true "id"
// @Tags     User
// @Router   /{id} [get]
func (s GrpcUserSvc) GetUser(c *fiber.Ctx) error {

	id := c.Params("id")
	user, err := s.repository.GetUserByID(id)
	if user == nil || err != nil {
		s.log.WithFields(log.Fields{"method": "GetUser"}).Errorf("No user found with %s \n", id)
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

	user := new(model.User)
	if err := c.BodyParser(user); err != nil {
		s.log.WithFields(log.Fields{"method": "CreateUser"}).Errorf("Review your input %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message:    fmt.Sprintf("Review your input %s", err),
			StatusCode: 400,
		})
	}

	err := s.validate.Struct(user)

	if err != nil {
		s.log.WithFields(log.Fields{"method": "CreateUser"}).Errorf("Review your payload %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message:    fmt.Sprintf("Review your input %s", err),
			StatusCode: 400,
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
		EventName:     pb.EventName_USER_CREATED,
	}
	err = handleEvent.Send(event)
	if err != nil {
		if err != nil {
			s.log.WithFields(log.Fields{"method": "CreateUser"}).Errorf("An error occured  %s \n", err)
			return c.Status(fiber.StatusBadRequest).JSON(&model.Response{
				Message:    err.Error(),
				StatusCode: 400,
			})
		}
	}

	recv, err := handleEvent.Recv()
	if err != nil {
		s.log.WithFields(log.Fields{"method": "CreateUser"}).Errorf("An error occured  %s \n", err)
		return c.Status(fiber.StatusBadRequest).JSON(&model.Response{
			Message:    err.Error(),
			StatusCode: 400,
		})
	} else {

		if recv.StatusCode >= 400 {
			s.log.WithFields(log.Fields{"method": "CreateUser"}).Errorf("An error occured %s \n", string(recv.Data))
			return c.Status(int(recv.StatusCode)).JSON(&model.Response{
				Data:       string(recv.Data),
				Message:    recv.Message,
				StatusCode: int(recv.StatusCode),
			})
		}

		var user model.User
		err = json.Unmarshal(recv.Data, &user)
		if err != nil {
			s.log.WithFields(log.Fields{"method": "CreateUser"}).Errorf("Couldn't unmarshall to recv.Data %s \n", err)
			return c.Status(fiber.StatusBadRequest).JSON(&model.Response{
				Message:    err.Error(),
				StatusCode: 400,
			})
		}

		s.log.WithFields(log.Fields{"method": "CreateUser"}).Infof("User created %v \n", user)
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
func (s GrpcUserSvc) UpdateUser(c *fiber.Ctx) error {

	user := new(model.User)
	if err := c.BodyParser(user); err != nil {
		s.log.WithFields(log.Fields{"method": "CreateUser"}).Errorf("Review your input %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message:    fmt.Sprintf("Review your input %s", err),
			StatusCode: 400,
		})
	}

	id := c.Params("id")
	if id == "" {
		s.log.WithFields(log.Fields{"method": "UpdateUser"}).Errorf("Review your query param %s \n", id)
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
	err = handleEvent.Send(event)
	if err != nil {
		if err != nil {
			s.log.WithFields(log.Fields{"method": "UpdateUser"}).Errorf("An error occured  %s \n", err)
			return c.Status(fiber.StatusBadRequest).JSON(&model.Response{
				Message:    err.Error(),
				StatusCode: 400,
			})
		}
	}

	recv, err := handleEvent.Recv()
	if err != nil {
		s.log.WithFields(log.Fields{"method": "UpdateUser"}).Errorf("An error occured  %s \n", err)
		return c.Status(fiber.StatusBadRequest).JSON(&model.Response{
			Message:    err.Error(),
			StatusCode: 400,
		})
	} else {

		if recv.StatusCode >= 400 {
			s.log.WithFields(log.Fields{"method": "UpdateUser"}).Errorf("An error occured %s \n", string(recv.Data))

			if string(recv.Data) != "" {
				return c.Status(int(recv.StatusCode)).JSON(&model.Response{
					Data:       string(recv.Data),
					Message:    recv.Message,
					StatusCode: int(recv.StatusCode),
				})
			} else {
				return c.Status(int(recv.StatusCode)).JSON(&model.Response{
					Message:    recv.Message,
					StatusCode: int(recv.StatusCode),
				})
			}
		}

		s.log.WithFields(log.Fields{"method": "UpdateUser"}).Infof("User successfully updated \n")
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
func (s GrpcUserSvc) DeleteUser(c *fiber.Ctx) error {

	id := c.Params("id")

	handleEvent, err := s.grpcClient.HandleEvent(context.Background())
	if err != nil {
		s.log.WithFields(log.Fields{"method": "DeleteUser"}).Errorf("An error occured \n")
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
	err = handleEvent.Send(event)
	if err != nil {
		s.log.WithFields(log.Fields{"method": "DeleteUser"}).Errorf("An error occured  %s \n", err)
		return c.Status(fiber.StatusBadRequest).JSON(&model.Response{
			Message:    err.Error(),
			StatusCode: 400,
		})
	}

	recv, err := handleEvent.Recv()
	if err != nil {
		s.log.WithFields(log.Fields{"method": "DeleteUser"}).Errorf("An error occured  %s \n", err)
		return c.Status(fiber.StatusBadRequest).JSON(&model.Response{
			Message:    err.Error(),
			StatusCode: 400,
		})
	} else {

		if recv.StatusCode >= 400 {
			s.log.WithFields(log.Fields{"method": "DeleteUser"}).Errorf("An error occured %s \n", string(recv.Data))
			if string(recv.Data) != "" {
				return c.Status(int(recv.StatusCode)).JSON(&model.Response{
					Data:       string(recv.Data),
					Message:    recv.Message,
					StatusCode: int(recv.StatusCode),
				})
			} else {
				return c.Status(int(recv.StatusCode)).JSON(&model.Response{
					Message:    recv.Message,
					StatusCode: int(recv.StatusCode),
				})
			}
		}

		s.log.WithFields(log.Fields{"method": "DeleteUser"}).Errorf("User successfully deleted %v \n", id)
		return c.Status(fiber.StatusOK).JSON(&model.Response{
			Message:    "User successfully deleted!",
			StatusCode: 200,
		})
	}

}

func NewGrpcUserService(rep repo.GrpcUserRepository, validate *validator.Validate, grpcClient pb.EventGrpcServiceClient, log *log.Entry, configs *user.AppConfig) GrpcUserService {
	return &GrpcUserSvc{
		repository: rep,
		validate:   validate,
		grpcClient: grpcClient,
		log:        log,
		configs:    configs,
	}
}
