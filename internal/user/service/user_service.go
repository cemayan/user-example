package service

import (
	"fmt"
	"github.com/cemayan/faceit-technical-test/config/user"
	"github.com/cemayan/faceit-technical-test/internal/user/dto"
	"github.com/cemayan/faceit-technical-test/internal/user/model"
	"github.com/cemayan/faceit-technical-test/internal/user/repo"
	"github.com/cemayan/faceit-technical-test/pkg/common"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	HashPassword(password string) (string, error)
	GetAllUser(c *fiber.Ctx) error
	GetUser(c *fiber.Ctx) error
	CreateUser(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
	HealthCheck(c *fiber.Ctx) error
}

// A UserSvc  contains the required dependencies for this service
type UserSvc struct {
	repository repo.UserRepository
	validate   *validator.Validate
	log        *log.Entry
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
func (s UserSvc) GetAllUser(c *fiber.Ctx) error {
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
func (s UserSvc) HealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("UP!")
}

// HashPassword returns encrypted password based on given password
func (s UserSvc) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// GetUser returns user based on given id.
// @Summary  GetUser
// @Param    id path string true "id"
// @Tags     User
// @Router   /{id} [get]
func (s UserSvc) GetUser(c *fiber.Ctx) error {

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
func (s UserSvc) CreateUser(c *fiber.Ctx) error {

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

	hash, err := s.HashPassword(user.Password)
	if err != nil {
		s.log.WithFields(log.Fields{"method": "CreateUser"}).Errorf("Couldn't hash password %s \n", err)
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message:    fmt.Sprintf("Couldn't hash password %s", err),
			StatusCode: 400,
		})
	}

	user.Password = hash

	user, err = s.repository.CreateUser(user)
	if user == nil || err != nil {
		s.log.WithFields(log.Fields{"method": "CreateUser"}).Errorf("Couldn't create use %s \n", err)
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message:    fmt.Sprintf("Couldn't create use %s", err),
			StatusCode: 400,
		})
	}

	newUser := dto.NewUser{
		Email:    user.Email,
		Nickname: user.NickName,
	}

	s.log.Infof("User created %s \n", newUser)
	return c.Status(fiber.StatusCreated).JSON(model.Response{
		StatusCode: 201,
		Data: model.UserData{
			ID:        user.ID,
			NickName:  user.NickName,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Country:   user.Country,
		},
		Message: fmt.Sprintf("User created %s", newUser),
	})

}

// UpdateUser return updated user based on given payload
// @Summary  UpdateUser
// @Param    id      path string         true "id"
// @Param    request body dto.UpdateUser true "query params"
// @Tags     User
// @Router   /{id} [put]
func (s UserSvc) UpdateUser(c *fiber.Ctx) error {

	var userDTO dto.UpdateUser
	if err := c.BodyParser(&userDTO); err != nil {
		s.log.WithFields(log.Fields{"method": "UpdateUser"}).Errorf("Review your input %s \n", err)
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			StatusCode: 400,
			Message:    fmt.Sprintf("Review your input %s", err),
		})
	}

	id := c.Params("id")
	if id == "" {
		s.log.WithFields(log.Fields{"method": "UpdateUser"}).Errorf("Review your id %v \n", id)
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			StatusCode: 400,
			Message:    fmt.Sprintf("Review your id %v", id),
		})
	}

	user, err := s.repository.GetUserByID(id)
	if user == nil || err != nil {
		s.log.WithFields(log.Fields{"method": "UpdateUser"}).Errorf("No user found with %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			StatusCode: 400,
			Message:    fmt.Sprintf("No user found with %s", err),
		})
	}

	if userDTO.Password != "" {
		hash, _ := s.HashPassword(userDTO.Password)
		user.Password = hash
	}
	if userDTO.NickName != "" {
		user.NickName = userDTO.NickName
	}
	if userDTO.Email != "" {
		user.Email = userDTO.Email
	}
	if userDTO.FirstName != "" {
		user.FirstName = userDTO.FirstName
	}
	if userDTO.LastName != "" {
		user.FirstName = userDTO.FirstName
	}
	if userDTO.Country != "" {
		user.Country = userDTO.Country
	}

	err = s.repository.UpdateUser(user)
	if err != nil {
		s.log.WithFields(log.Fields{"method": "UpdateUser"}).Errorf("While user is updating an error occured: %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			StatusCode: 400,
			Message:    fmt.Sprintf("While user is updating an error occured: %s", err),
		})
	} else {
		s.log.WithFields(log.Fields{"method": "UpdateUser"}).Infof("User successfully updated \n")
		return c.Status(fiber.StatusOK).JSON(model.Response{
			StatusCode: 200,
			Data: model.UserData{
				ID:        user.ID,
				NickName:  user.NickName,
				Email:     user.Email,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Country:   user.Country,
			},
			Message: "User successfully updated",
		})
	}
}

// DeleteUser removes  the user based on given payload
// @Summary  DeleteUser
// @Param    id path string true "id"
// @Tags     User
// @Router   /{id} [delete]
func (s UserSvc) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	err := s.repository.DeleteUser(id)
	if err != nil {

		s.log.WithFields(log.Fields{"method": "CreateUser"}).Errorf("While user is deleting an error occured: %s \n", err)
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			StatusCode: 400,
			Message:    fmt.Sprintf("While user is deleting an error occured: %s", err),
		})
	} else {
		s.log.WithFields(log.Fields{"method": "DeleteUser"}).Errorf("User successfully deleted %v \n", id)
		return c.Status(fiber.StatusOK).JSON(model.Response{
			StatusCode: 200,
			Message:    fmt.Sprintf("User successfully deleted %v", id),
		})
	}
}

func NewUserService(rep repo.UserRepository, validate *validator.Validate, log *log.Entry, configs *user.AppConfig) UserService {
	return &UserSvc{
		repository: rep,
		validate:   validate,
		log:        log,
		configs:    configs,
	}
}
