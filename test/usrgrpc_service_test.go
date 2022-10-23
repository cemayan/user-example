package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cemayan/faceit-technical-test/config/user"

	"github.com/cemayan/faceit-technical-test/internal/usrgrpc/dto"
	"github.com/cemayan/faceit-technical-test/internal/usrgrpc/model"
	"github.com/cemayan/faceit-technical-test/internal/usrgrpc/repo"
	"github.com/cemayan/faceit-technical-test/internal/usrgrpc/router"
	"github.com/cemayan/faceit-technical-test/internal/usrgrpc/service"
	"github.com/cemayan/faceit-technical-test/internal/usrgrpc/util"
	"github.com/cemayan/faceit-technical-test/pkg/postgres"
	pb "github.com/cemayan/faceit-technical-test/protos/event"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
	"io"
	"net"
	"net/http/httptest"
	"os"
	"testing"
)

var userRepo repo.GrpcUserRepository

type e2eGrpcTestSuite struct {
	suite.Suite
	app       *fiber.App
	db        *gorm.DB
	usrSvc    service.GrpcUserService
	configs   *user.AppConfig
	v         *viper.Viper
	dbHandler postgres.DBHandler
	validate  *validator.Validate
}

func TestE2EGrpcTestSuite(t *testing.T) {
	suite.Run(t, &e2eGrpcTestSuite{})
}

func (ts *e2eGrpcTestSuite) SetupSuite() {

	app := fiber.New()
	app.Use(cors.New())

	ts.app = app

	ts.v = viper.New()
	_configs := user.NewConfig(ts.v)

	ts.validate = validator.New()

	env := os.Getenv("ENV")
	appConfig, err := _configs.GetConfig(env)
	ts.configs = appConfig
	if err != nil {
		return
	}

	//Postresql connection
	ts.dbHandler = postgres.NewDBHandler(&ts.configs.Postgresql, log.New().WithFields(log.Fields{"service": "user_grpc"}))
	db := ts.dbHandler.New()
	ts.db = db

	_userRepo := repo.NewGrpcUserRepo(ts.db, log.New().WithFields(log.Fields{"service": "user"}))
	userRepo = _userRepo

	util.MigrateDB(ts.db, log.New().WithFields(log.Fields{"service": "user"}))

	str := fmt.Sprintf("%s:%s", ts.configs.Grpc.ADDR, ts.configs.Grpc.PORT)
	//gRPC connection
	_grpcConn, err := grpc.Dial(str, grpc.WithTransportCredentials(insecure.NewCredentials()))
	util.FailOnError(err, "did not connect")

	grpcClient := pb.NewEventGrpcServiceClient(_grpcConn)

	userSvc := service.NewGrpcUserService(_userRepo, ts.validate, grpcClient, log.New().WithFields(log.Fields{"service": "user_grpc"}), ts.configs)
	ts.usrSvc = userSvc

	log.Infoln("gRPC server is starting...")
	_, err = net.Listen("tcp", fmt.Sprintf(":%s", ts.configs.Grpc.PORT))
	util.FailOnError(err, "tcp listen failed.")

	router.SetupGrpcRoutes(ts.app, log.New().WithFields(log.Fields{"service": "user"}), grpcClient, appConfig)

}

func (ts *e2eGrpcTestSuite) removeAllRecords() {
	ts.db.Exec("DELETE FROM users")
}

func (ts *e2eGrpcTestSuite) getRecords() []model.User {
	var users []model.User
	ts.db.Find(&users)
	return users
}

func (ts *e2eGrpcTestSuite) getUserModel() model.User {
	var user model.User
	user.Password = "123"
	user.NickName = "test"
	user.Email = "user@test.com"
	user.Country = "UK"
	return user
}

func (ts *e2eGrpcTestSuite) getWrongUserModel() model.User {
	var user model.User
	user.Country = "UK"
	return user
}

func (ts *e2eGrpcTestSuite) getUpdateUserModel() dto.UpdateUser {
	var user dto.UpdateUser
	user.NickName = "test4"
	user.Email = "user@test.com"
	return user
}

func (ts *e2eGrpcTestSuite) saveUserModel() model.User {
	var user model.User

	password, _ := ts.usrSvc.HashPassword("123")
	user.Password = password
	user.NickName = "test"
	user.Email = "user@test.com"
	user.Country = "UK"
	ts.db.Create(&user)
	return user
}

func (ts *e2eGrpcTestSuite) TestUserService_Create() {

	ts.removeAllRecords()

	ts.app.Post("/user", ts.usrSvc.CreateUser)

	marshal, err := json.Marshal(ts.getUserModel())
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("POST", "/user", requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	_, err = ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	users := ts.getRecords()

	if len(users) > 0 {
		ts.Equal("test", users[0].NickName)
		ts.Equal("user@test.com", users[0].Email)
	}

}

func (ts *e2eGrpcTestSuite) TestUserService_CreateWrongPayload() {

	ts.removeAllRecords()

	ts.app.Post("/user", ts.usrSvc.CreateUser)

	marshal, err := json.Marshal("test")
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("POST", "/user", requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

func (ts *e2eGrpcTestSuite) TestUserService_CreateWrongPayload2() {

	ts.removeAllRecords()

	ts.app.Post("/user", ts.usrSvc.CreateUser)

	marshal, err := json.Marshal(ts.getWrongUserModel())
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("POST", "/user", requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

func (ts *e2eGrpcTestSuite) TestUserService_CheckStatusCodeWhenCreate() {

	ts.removeAllRecords()

	ts.app.Post("/user", ts.usrSvc.CreateUser)

	marshal, err := json.Marshal(ts.getUserModel())
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("POST", "/user", requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}
	fmt.Println(resp)
	ts.Equal(fiber.StatusCreated, resp.StatusCode)
}

func (ts *e2eGrpcTestSuite) TestUserService_SameEmail() {

	ts.removeAllRecords()
	ts.saveUserModel()

	ts.app.Post("/user", ts.usrSvc.CreateUser)

	marshal, err := json.Marshal(ts.getUserModel())
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("POST", "/user", requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

func (ts *e2eGrpcTestSuite) TestUserService_SameNickname() {

	ts.removeAllRecords()
	ts.saveUserModel()

	ts.app.Post("/user", ts.usrSvc.CreateUser)

	marshal, err := json.Marshal(ts.getUserModel())
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("POST", "/user", requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

func (ts *e2eGrpcTestSuite) TestUserService_UpdateUser() {

	ts.removeAllRecords()
	userModel := ts.saveUserModel()

	ts.app.Put("/user/:id", ts.usrSvc.UpdateUser)

	marshal, err := json.Marshal(ts.getUpdateUserModel())
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("PUT", "/user/"+userModel.ID.String(), requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	users := ts.getRecords()

	if len(users) > 0 {
		ts.Equal(fiber.StatusOK, resp.StatusCode)
		ts.Equal("test4", users[0].NickName)

	}

}

func (ts *e2eGrpcTestSuite) TestUserService_UpdateWrongPayload() {

	ts.removeAllRecords()
	userModel := ts.saveUserModel()

	ts.app.Put("/user", ts.usrSvc.CreateUser)

	marshal, err := json.Marshal("test")
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("PUT", "/user/"+userModel.ID.String(), requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

func (ts *e2eGrpcTestSuite) TestUserService_UpdateWrongId() {

	ts.removeAllRecords()
	ts.saveUserModel()

	ts.app.Put("/user", ts.usrSvc.CreateUser)

	marshal, err := json.Marshal("test")
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("PUT", "/user/312342354642", requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

func (ts *e2eGrpcTestSuite) TestUserService_UpdateWrongUser() {

	ts.removeAllRecords()

	ts.app.Put("/user/:id", ts.usrSvc.UpdateUser)

	marshal, err := json.Marshal(ts.getUpdateUserModel())
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("PUT", "/user/12312312312312", requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

func (ts *e2eGrpcTestSuite) TestUserService_DeleteUser() {

	ts.removeAllRecords()
	userModel := ts.saveUserModel()

	ts.app.Delete("/user/:id", ts.usrSvc.DeleteUser)

	marshal, err := json.Marshal(ts.getUpdateUserModel())
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("DELETE", "/user/"+userModel.ID.String(), requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	users := ts.getRecords()

	ts.Equal(fiber.StatusOK, resp.StatusCode)
	ts.Equal(0, len(users))

}

func (ts *e2eGrpcTestSuite) TestUserService_DeleteWrongUser() {

	ts.removeAllRecords()

	ts.app.Delete("/user/:id", ts.usrSvc.DeleteUser)

	marshal, err := json.Marshal(ts.getUpdateUserModel())
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("DELETE", "/user/3123213", requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusBadRequest, resp.StatusCode)

}

func (ts *e2eGrpcTestSuite) TestUserService_GetUser() {

	ts.removeAllRecords()
	userModel := ts.saveUserModel()

	ts.app.Get("/user/:id", ts.usrSvc.GetUser)

	req := httptest.NewRequest("GET", "/user/"+userModel.ID.String(), nil)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	answer, _ := io.ReadAll(resp.Body)

	var response model.Response
	err = json.Unmarshal(answer, &response)
	if err != nil {
		return
	}

	var user model.User
	str, err := json.Marshal(response.Data)
	if err != nil {
		return
	}

	err = json.Unmarshal(str, &user)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusOK, resp.StatusCode)
	ts.Equal("test", user.NickName)

}

func (ts *e2eGrpcTestSuite) TestUserService_GetUserWrongId() {

	ts.removeAllRecords()
	ts.saveUserModel()

	ts.app.Get("/user/:id", ts.usrSvc.GetUser)

	req := httptest.NewRequest("GET", "/user/34253623564", nil)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	answer, _ := io.ReadAll(resp.Body)

	var response model.Response
	err = json.Unmarshal(answer, &response)
	if err != nil {
		return
	}

	var user model.User
	str, err := json.Marshal(response.Data)
	if err != nil {
		return
	}

	err = json.Unmarshal(str, &user)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusBadRequest, resp.StatusCode)

}

func (ts *e2eGrpcTestSuite) TestUserService_GetAllUser() {

	ts.removeAllRecords()
	ts.saveUserModel()

	ts.app.Get("/user/", ts.usrSvc.GetAllUser)

	req := httptest.NewRequest("GET", "/user/", nil)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	users := ts.getRecords()
	ts.Equal(fiber.StatusOK, resp.StatusCode)
	ts.Equal(1, len(users))

}

func (ts *e2eGrpcTestSuite) TestUserService_GetAllUserWrongQueryParams() {

	ts.removeAllRecords()
	ts.saveUserModel()

	ts.app.Get("/user/", ts.usrSvc.GetAllUser)

	req := httptest.NewRequest("GET", "/user/?limit=abc", nil)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusBadRequest, resp.StatusCode)

}

func (ts *e2eGrpcTestSuite) TestUserService_GetAllUserWithParams() {

	ts.removeAllRecords()
	ts.saveUserModel()

	ts.app.Get("/user/", ts.usrSvc.GetAllUser)

	req := httptest.NewRequest("GET", "/user/?limit=1&page=1", nil)
	req.Header.Add("Content-Type", "application/json")

	_, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	users := ts.getRecords()

	ts.Equal(1, len(users))

}

func (ts *e2eGrpcTestSuite) TestUserService_GetAllUserWithWrongParams() {

	ts.removeAllRecords()
	ts.saveUserModel()

	ts.app.Get("/user/", ts.usrSvc.GetAllUser)

	req := httptest.NewRequest("GET", "/user/?limit=dasda&page=dadas", nil)
	req.Header.Add("Content-Type", "application/json")

	_, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	users := ts.getRecords()

	ts.Equal(1, len(users))

}

func (ts *e2eGrpcTestSuite) TestUserService_GetAllUserWithParams2() {

	ts.removeAllRecords()
	ts.saveUserModel()

	ts.app.Get("/user/", ts.usrSvc.GetAllUser)

	req := httptest.NewRequest("GET", "/user/?limit=1&page=1&cQuery=country%20%3D%20%3F&cValue=UK", nil)
	req.Header.Add("Content-Type", "application/json")

	_, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	users := ts.getRecords()

	ts.Equal(1, len(users))

}
