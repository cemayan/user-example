package test

import (
	"bytes"
	"encoding/json"
	"github.com/cemayan/faceit-technical-test/config/user"
	"github.com/cemayan/faceit-technical-test/internal/user/model"
	"github.com/cemayan/faceit-technical-test/internal/user/repo"
	"github.com/cemayan/faceit-technical-test/internal/user/router"
	"github.com/cemayan/faceit-technical-test/internal/user/service"
	"github.com/cemayan/faceit-technical-test/internal/user/util"
	"github.com/cemayan/faceit-technical-test/internal/usrgrpc/dto"
	"github.com/cemayan/faceit-technical-test/pkg/postgres"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"io"
	"net/http/httptest"
	"os"
	"testing"
)

type e2eTestSuite struct {
	suite.Suite
	app       *fiber.App
	db        *gorm.DB
	usrSvc    service.UserService
	configs   *user.AppConfig
	v         *viper.Viper
	dbHandler postgres.DBHandler
	validate  *validator.Validate
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, &e2eTestSuite{})
}

func (ts *e2eTestSuite) SetupSuite() {

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

	router.SetupRoutes(ts.app, log.New().WithFields(log.Fields{"service": "user"}), appConfig)

	//Postresql connection
	ts.dbHandler = postgres.NewDBHandler(&ts.configs.Postgresql, log.New().WithFields(log.Fields{"service": "user"}))
	db := ts.dbHandler.New()
	ts.db = db

	util.MigrateDB(ts.db, log.New().WithFields(log.Fields{"service": "user"}))

	userRepo := repo.NewUserRepo(ts.db, log.New().WithFields(log.Fields{"service": "user"}))

	userSvc := service.NewUserService(userRepo, ts.validate, log.New().WithFields(log.Fields{"service": "user"}), ts.configs)
	ts.usrSvc = userSvc

}

func (ts *e2eTestSuite) removeAllRecords() {
	ts.db.Exec("DELETE FROM users")
}

func (ts *e2eTestSuite) getRecords() []model.User {
	var users []model.User
	ts.db.Find(&users)
	return users
}

func (ts *e2eTestSuite) getUserModel() model.User {
	var userModel model.User
	userModel.Password = "123"
	userModel.NickName = "test"
	userModel.Email = "user@test.com"
	userModel.Country = "UK"
	return userModel
}

func (ts *e2eTestSuite) getWrongUserModel() model.User {
	var userModel model.User
	userModel.Country = "UK"
	return userModel
}

func (ts *e2eTestSuite) getUpdateUserModel() dto.UpdateUser {
	var userModel dto.UpdateUser
	userModel.NickName = "test4"
	userModel.Email = "user@test.com"
	return userModel
}

func (ts *e2eTestSuite) saveUserModel() model.User {
	var userModel model.User

	password, _ := ts.usrSvc.HashPassword("123")
	userModel.Password = password
	userModel.NickName = "test"
	userModel.Email = "user@test.com"
	userModel.Country = "UK"
	ts.db.Create(&userModel)
	return userModel
}

func (ts *e2eTestSuite) TestUserService_Create() {

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

func (ts *e2eTestSuite) TestUserService_CreateWrongPayload() {

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

func (ts *e2eTestSuite) TestUserService_CreateWrongPayload2() {

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

func (ts *e2eTestSuite) TestUserService_CheckStatusCodeWhenCreate() {

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
	ts.Equal(fiber.StatusCreated, resp.StatusCode)
}

func (ts *e2eTestSuite) TestUserService_SameEmail() {

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

func (ts *e2eTestSuite) TestUserService_SameNickname() {

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

func (ts *e2eTestSuite) TestUserService_UpdateUser() {

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

func (ts *e2eTestSuite) TestUserService_UpdateWrongPayload() {

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

func (ts *e2eTestSuite) TestUserService_UpdateWrongId() {

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

func (ts *e2eTestSuite) TestUserService_UpdateWrongUser() {

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

func (ts *e2eTestSuite) TestUserService_DeleteUser() {

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

func (ts *e2eTestSuite) TestUserService_DeleteWrongUser() {

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

func (ts *e2eTestSuite) TestUserService_GetUser() {

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

	var userM model.User
	str, err := json.Marshal(response.Data)
	if err != nil {
		return
	}

	err = json.Unmarshal(str, &userM)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusOK, resp.StatusCode)
	ts.Equal("test", userM.NickName)

}

func (ts *e2eTestSuite) TestUserService_GetUserWrongId() {

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

	var userModel model.User
	str, err := json.Marshal(response.Data)
	if err != nil {
		return
	}

	err = json.Unmarshal(str, &userModel)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusBadRequest, resp.StatusCode)

}

func (ts *e2eTestSuite) TestUserService_GetAllUser() {

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

func (ts *e2eTestSuite) TestUserService_GetAllUserWrongQueryParams() {

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

func (ts *e2eTestSuite) TestUserService_GetAllUserWithParams() {

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

func (ts *e2eTestSuite) TestUserService_GetAllUserWithWrongParams() {

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

func (ts *e2eTestSuite) TestUserService_GetAllUserWithParams2() {

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
