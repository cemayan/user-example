package service

import (
	"fmt"
	"github.com/cemayan/faceit-technical-test/config/user"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/model"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/repo"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type GrpcAuthService interface {
	isValidUser(id string) bool
	isValidUserId(t *jwt.Token, id string) bool
	HealthCheck(c *fiber.Ctx) error
	getUserByEmail(e string) (*model.User, error)
	getUserByNickname(u string) (*model.User, error)
	Login(c *fiber.Ctx) error
}

// A AuthSvc  contains the required dependencies for this service
type GrpcAuthSvc struct {
	repository repo.GrpcUserRepository
	log        *log.Logger
	configs    *user.AppConfig
}

func (s GrpcAuthSvc) isValidUser(id string) bool {

	user, err := s.repository.GetUserById(id)
	if user == nil || err != nil {
		return false
	}
	return true
}

// validToken  returns valid status based on given token
// If based on given token is not found in claims map, it is returned false
func (s GrpcAuthSvc) isValidUserId(t *jwt.Token, id string) bool {

	claims := t.Claims.(jwt.MapClaims)
	uid := claims["user_id"]

	if uid != id {
		return false
	}

	return true
}

// HealthCheck returns 200 with body
func (s GrpcAuthSvc) HealthCheck(c *fiber.Ctx) error {
	return c.Status(200).JSON("healty!")
}

// CheckPasswordHash returns the correctness of password
func (s GrpcAuthSvc) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CheckPasswordHash returns user  based on given email.
func (s GrpcAuthSvc) getUserByEmail(e string) (*model.User, error) {
	return s.repository.GetUserByEmail(e)
}

// CheckPasswordHash returns user  based on given username.
func (s GrpcAuthSvc) getUserByNickname(u string) (*model.User, error) {
	return s.repository.GetUserByNickname(u)
}

// Login returns authentication result
// If given password or username is not correct, it is returned 403
// Then, it is created new jwt token. Nickname, email , user_id and exp is added to token claims.
// @Summary Login
// @Param   request body model.LoginInput true "query params"
// @Tags    Auth
// @Router  /getToken [post]
func (s GrpcAuthSvc) Login(c *fiber.Ctx) error {

	var input model.LoginInput
	var ud model.UserData

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message:    fmt.Sprintf("Error on login request %v", err),
			StatusCode: 400,
		})
	}
	identity := input.Nickname
	pass := input.Password

	user, err := s.getUserByNickname(identity)
	if user == nil || err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
			Message:    fmt.Sprintf("Error on username %v", err),
			StatusCode: 400,
		})
	}

	ud = model.UserData{
		NickName: user.NickName,
		Email:    user.Email,
		Password: user.Password,
	}

	if !s.CheckPasswordHash(pass, ud.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
			Message:    fmt.Sprintf("Invalid password %v", identity),
			StatusCode: 400,
		})
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["nick_name"] = ud.NickName
	claims["email"] = ud.Email
	claims["user_id"] = user.ID.String()
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	t, err := token.SignedString([]byte(s.configs.SECRET))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Data:       fmt.Sprintf("Invalid  secret %v", nil),
			StatusCode: 500,
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.Response{
		Data: model.UserData{
			ID:        user.ID,
			NickName:  user.NickName,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Country:   user.Country,
		},
		Message:    fmt.Sprintf("%v", t),
		StatusCode: 200,
	})

}

func NewAuthService(rep repo.GrpcUserRepository, log *log.Logger, configs *user.AppConfig) GrpcAuthService {
	return &GrpcAuthSvc{
		repository: rep,
		log:        log,
		configs:    configs,
	}
}
