package repo

import (
	"errors"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/dto"
	"github.com/cemayan/faceit-technical-test/internal/user_grpc/model"
	"github.com/cemayan/faceit-technical-test/pkg/common"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"math"
)

type GrpcUserRepository interface {
	GetAllUser(pagination common.Pagination) (*common.Pagination, error)
	CreateUser(user *model.User) (*model.User, error)
	UpdateUser(id string, user *dto.UpdateUser) error
	DeleteUser(id string) error
	GetUserById(id string) (*model.User, error)
	hashPassword(password string) (string, error)
	paginate(value interface{}, pagination *common.Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB
}

type GrpcUserrepo struct {
	db  *gorm.DB
	log *log.Entry
}

func (r GrpcUserrepo) paginate(value interface{}, pagination *common.Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	db.Model(value).Count(&totalRows)

	pagination.TotalRows = totalRows
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))
	pagination.TotalPages = totalPages

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}

func (r GrpcUserrepo) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (r GrpcUserrepo) GetAllUser(pagination common.Pagination) (*common.Pagination, error) {
	var users []model.User

	if pagination.CQuery != "" && pagination.CValue != "" {
		conditionQuery := pagination.CQuery
		conditionValue := pagination.CValue
		tx := r.db.Scopes(r.paginate(users, &pagination, r.db)).Debug().Where(conditionQuery, conditionValue).Find(&users)
		pagination.Rows = users
		return &pagination, tx.Error
	} else if pagination.CQuery == "" && pagination.CValue == "" {
		tx := r.db.Scopes(r.paginate(users, &pagination, r.db)).Find(&users)
		pagination.Rows = users
		return &pagination, tx.Error
	} else {
		return nil, nil
	}

}

func (r GrpcUserrepo) UpdateUser(id string, userDTO *dto.UpdateUser) error {

	user, err := r.GetUserById(id)

	if user == nil || err != nil {
		return err
	}

	if userDTO.Password != "" {
		hash, _ := r.hashPassword(userDTO.Password)
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

	tx := r.db.Save(user)
	return tx.Error
}

func (r GrpcUserrepo) DeleteUser(id string) error {
	user, err := r.GetUserById(id)

	if err != nil {
		return err
	}

	tx := r.db.Delete(user)
	return tx.Error
}

func (r GrpcUserrepo) CreateUser(user *model.User) (*model.User, error) {

	hash, err := r.hashPassword(user.Password)
	if err != nil {
		return nil, err
	}

	user.Password = hash

	if err := r.db.Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return user, nil
}

// GetUserById returns user based on given id
func (r GrpcUserrepo) GetUserById(id string) (*model.User, error) {
	var user model.User
	_id, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	if err := r.db.Where("id = ?", _id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}

	return &user, nil
}

func NewGrpcUserRepo(db *gorm.DB, log *log.Entry) GrpcUserRepository {
	return &GrpcUserrepo{
		db:  db,
		log: log,
	}
}
