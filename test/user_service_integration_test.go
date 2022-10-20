package test

import (
	"bytes"
	"encoding/json"
	"github.com/cemayan/faceit-technical-test/config/user"
	"github.com/cemayan/faceit-technical-test/internal/user/middleware"
	"github.com/cemayan/faceit-technical-test/internal/user/model"
	"github.com/cemayan/faceit-technical-test/internal/user/repo"
	"github.com/cemayan/faceit-technical-test/internal/user/router"
	"github.com/cemayan/faceit-technical-test/internal/user/service"
	"github.com/cemayan/faceit-technical-test/internal/user/util"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/dto"
	"github.com/cemayan/faceit-technical-test/pkg/postgres"
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
	authSvc   service.AuthService
	configs   *user.AppConfig
	v         *viper.Viper
	dbHandler postgres.DBHandler
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

	env := os.Getenv("ENV")
	appConfig, err := _configs.GetConfig(env)
	ts.configs = appConfig
	if err != nil {
		return
	}

	router.SetupRoutes(ts.app, log.New(), appConfig)

	//Postresql connection
	ts.dbHandler = postgres.NewDbHandler(&ts.configs.Postgresql)
	db := ts.dbHandler.New()
	ts.db = db

	util.MigrateDB(ts.db)

	userRepo := repo.NewUserRepo(ts.db, log.New())
	authSvc := service.NewAuthService(userRepo, log.New(), ts.configs)
	ts.authSvc = authSvc

	userSvc := service.NewUserService(userRepo, authSvc, log.New(), ts.configs)
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
	var user model.User
	user.Password = "123"
	user.NickName = "test"
	user.Email = "user@test.com"
	user.Country = "UK"
	return user
}

func (ts *e2eTestSuite) getUpdateUserModel() dto.UpdateUser {
	var user dto.UpdateUser
	user.NickName = "test4"
	user.Email = "user@test.com"
	return user
}

func (ts *e2eTestSuite) getAuthModel() model.LoginInput {
	var login model.LoginInput
	login.Nickname = "test"
	login.Password = "123"
	return login
}

func (ts *e2eTestSuite) getWrongAuthModel() model.LoginInput {
	var login model.LoginInput
	login.Nickname = "test"
	login.Password = "1234"
	return login
}

func (ts *e2eTestSuite) getToken() string {

	ts.app.Post("/getToken", ts.authSvc.Login)

	marshal, err := json.Marshal(ts.getAuthModel())
	if err != nil {
		return ""
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("POST", "/getToken", requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return ""
	}

	answer, err := io.ReadAll(resp.Body)

	var response model.Response
	err = json.Unmarshal(answer, &response)
	return response.Message
}

func (ts *e2eTestSuite) saveUserModel() model.User {
	var user model.User

	password, _ := ts.usrSvc.HashPassword("123")
	user.Password = password
	user.NickName = "test"
	user.Email = "user@test.com"
	user.Country = "UK"
	ts.db.Create(&user)
	return user
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

func (ts *e2eTestSuite) TestAuthService_SuccessLogin() {
	ts.removeAllRecords()
	ts.saveUserModel()

	ts.app.Post("/getToken", ts.authSvc.Login)
	marshal, err := json.Marshal(ts.getAuthModel())
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("POST", "/getToken", requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	answer, err := io.ReadAll(resp.Body)

	var response model.Response
	err = json.Unmarshal(answer, &response)

	ts.Equal(fiber.StatusOK, resp.StatusCode)
}

func (ts *e2eTestSuite) TestAuthService_FailedLogin() {
	ts.removeAllRecords()
	ts.saveUserModel()

	ts.app.Post("/getToken", ts.authSvc.Login)
	marshal, err := json.Marshal(ts.getWrongAuthModel())
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("POST", "/getToken", requestBuffer)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	answer, err := io.ReadAll(resp.Body)

	var response model.Response
	err = json.Unmarshal(answer, &response)

	ts.Equal(fiber.StatusUnauthorized, resp.StatusCode)
}

func (ts *e2eTestSuite) TestUserService_UpdateUser() {

	ts.removeAllRecords()
	userModel := ts.saveUserModel()

	token := ts.getToken()
	if token != "" {
		ts.app.Put("/user/:id", middleware.Protected(ts.configs), ts.usrSvc.UpdateUser)

		marshal, err := json.Marshal(ts.getUpdateUserModel())
		if err != nil {
			return
		}

		requestBuffer := bytes.NewBuffer(marshal)

		req := httptest.NewRequest("PUT", "/user/"+userModel.ID.String(), requestBuffer)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+token)

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

}

func (ts *e2eTestSuite) TestUserService_UpdateUserWhenWrongUser() {

	ts.removeAllRecords()
	ts.saveUserModel()

	token := ts.getToken()
	if token != "" {
		ts.app.Put("/user/:id", middleware.Protected(ts.configs), ts.usrSvc.UpdateUser)

		marshal, err := json.Marshal(ts.getUpdateUserModel())
		if err != nil {
			return
		}

		requestBuffer := bytes.NewBuffer(marshal)

		req := httptest.NewRequest("PUT", "/user/1231231", requestBuffer)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+token)

		resp, err := ts.app.Test(req, 10000)
		if err != nil {
			return
		}

		ts.Equal(fiber.StatusUnauthorized, resp.StatusCode)

	}

}

func (ts *e2eTestSuite) TestUserService_UpdateUserWhenWrongToken() {

	ts.removeAllRecords()
	userModel := ts.saveUserModel()

	token := "test"

	ts.app.Put("/user/:id", middleware.Protected(ts.configs), ts.usrSvc.UpdateUser)

	marshal, err := json.Marshal(ts.getUpdateUserModel())
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("PUT", "/user/"+userModel.ID.String(), requestBuffer)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusUnauthorized, resp.StatusCode)

}

func (ts *e2eTestSuite) TestUserService_DeleteUser() {

	ts.removeAllRecords()
	userModel := ts.saveUserModel()

	token := ts.getToken()
	if token != "" {
		ts.app.Delete("/user/:id", middleware.Protected(ts.configs), ts.usrSvc.DeleteUser)

		marshal, err := json.Marshal(ts.getUpdateUserModel())
		if err != nil {
			return
		}

		requestBuffer := bytes.NewBuffer(marshal)

		req := httptest.NewRequest("DELETE", "/user/"+userModel.ID.String(), requestBuffer)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+token)

		resp, err := ts.app.Test(req, 10000)
		if err != nil {
			return
		}

		users := ts.getRecords()

		ts.Equal(fiber.StatusOK, resp.StatusCode)
		ts.Equal(0, len(users))

	}

}

func (ts *e2eTestSuite) TestUserService_DeleteUserWhenWrongUser() {

	ts.removeAllRecords()
	ts.saveUserModel()

	token := ts.getToken()
	if token != "" {
		ts.app.Delete("/user/:id", middleware.Protected(ts.configs), ts.usrSvc.DeleteUser)

		marshal, err := json.Marshal(ts.getUpdateUserModel())
		if err != nil {
			return
		}

		requestBuffer := bytes.NewBuffer(marshal)

		req := httptest.NewRequest("DELETE", "/user/adadasdas", requestBuffer)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+token)

		resp, err := ts.app.Test(req, 10000)
		if err != nil {
			return
		}

		ts.Equal(fiber.StatusUnauthorized, resp.StatusCode)

	}

}

func (ts *e2eTestSuite) TestUserService_DeleteUserWhenWrongToken() {

	ts.removeAllRecords()
	userModel := ts.saveUserModel()

	token := "test"
	ts.app.Delete("/user/:id", middleware.Protected(ts.configs), ts.usrSvc.DeleteUser)

	marshal, err := json.Marshal(ts.getUpdateUserModel())
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(marshal)

	req := httptest.NewRequest("DELETE", "/user/"+userModel.ID.String(), requestBuffer)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusUnauthorized, resp.StatusCode)

}

func (ts *e2eTestSuite) TestUserService_GetUser() {

	ts.removeAllRecords()
	userModel := ts.saveUserModel()

	token := ts.getToken()
	if token != "" {
		ts.app.Get("/user/:id", middleware.Protected(ts.configs), ts.usrSvc.GetUser)

		req := httptest.NewRequest("GET", "/user/"+userModel.ID.String(), nil)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+token)

		resp, err := ts.app.Test(req, 10000)
		if err != nil {
			return
		}

		answer, err := io.ReadAll(resp.Body)

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

}

func (ts *e2eTestSuite) TestUserService_GetUserWhenWrongToken() {

	ts.removeAllRecords()
	userModel := ts.saveUserModel()

	token := "test"

	ts.app.Get("/user/:id", middleware.Protected(ts.configs), ts.usrSvc.GetUser)

	req := httptest.NewRequest("GET", "/user/"+userModel.ID.String(), nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := ts.app.Test(req, 10000)
	if err != nil {
		return
	}

	answer, err := io.ReadAll(resp.Body)

	var response model.Response
	err = json.Unmarshal(answer, &response)
	if err != nil {
		return
	}

	ts.Equal(fiber.StatusUnauthorized, resp.StatusCode)

}

func (ts *e2eTestSuite) TestUserService_GetAllUser() {

	ts.removeAllRecords()
	ts.saveUserModel()

	token := ts.getToken()
	if token != "" {
		ts.app.Get("/user/", middleware.Protected(ts.configs), ts.usrSvc.GetAllUser)

		req := httptest.NewRequest("GET", "/user/", nil)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+token)

		resp, err := ts.app.Test(req, 10000)
		if err != nil {
			return
		}

		users := ts.getRecords()
		ts.Equal(fiber.StatusOK, resp.StatusCode)
		ts.Equal(1, len(users))
	}

}

func (ts *e2eTestSuite) TestUserService_GetAllUserWrongQueryParams() {

	ts.removeAllRecords()
	ts.saveUserModel()

	token := ts.getToken()
	if token != "" {
		ts.app.Get("/user/", middleware.Protected(ts.configs), ts.usrSvc.GetAllUser)

		req := httptest.NewRequest("GET", "/user/?limit=abc", nil)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+token)

		resp, err := ts.app.Test(req, 10000)
		if err != nil {
			return
		}

		ts.Equal(fiber.StatusBadRequest, resp.StatusCode)

	}

}

func (ts *e2eTestSuite) TestUserService_GetAllUserWithParams() {

	ts.removeAllRecords()
	ts.saveUserModel()

	token := ts.getToken()
	if token != "" {
		ts.app.Get("/user/", middleware.Protected(ts.configs), ts.usrSvc.GetAllUser)

		req := httptest.NewRequest("GET", "/user/?limit=1&page=1", nil)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+token)

		_, err := ts.app.Test(req, 10000)
		if err != nil {
			return
		}

		users := ts.getRecords()

		ts.Equal(1, len(users))

	}

}

func (ts *e2eTestSuite) TestUserService_GetAllUserWithParams2() {

	ts.removeAllRecords()
	ts.saveUserModel()

	token := ts.getToken()
	if token != "" {
		ts.app.Get("/user/", middleware.Protected(ts.configs), ts.usrSvc.GetAllUser)

		req := httptest.NewRequest("GET", "/user/?limit=1&page=1&cQuery=country%20%3D%20%3F&cValue=UK", nil)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+token)

		_, err := ts.app.Test(req, 10000)
		if err != nil {
			return
		}

		users := ts.getRecords()

		ts.Equal(1, len(users))
	}

}
