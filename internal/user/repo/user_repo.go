package repo

import (
	"errors"
	"github.com/cemayan/faceit-technical-test/internal/user/model"
	"github.com/cemayan/faceit-technical-test/pkg/common"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"math"
)

type UserRepository interface {
	GetAllUser(pagination common.Pagination) (*common.Pagination, error)
	CreateUser(user *model.User) (*model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(id string) error
	GetUserById(id string) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	GetUserByNickname(nickName string) (*model.User, error)
	paginate(value interface{}, pagination *common.Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB
}

type Userrepo struct {
	db  *gorm.DB
	log *log.Logger
}

func (r Userrepo) paginate(value interface{}, pagination *common.Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	db.Model(value).Count(&totalRows)

	pagination.TotalRows = totalRows
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))
	pagination.TotalPages = totalPages

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}

func (r Userrepo) GetAllUser(pagination common.Pagination) (*common.Pagination, error) {
	var users []model.User

	if pagination.CQuery != "" && pagination.CValue != "" {
		conditionQuery := pagination.CQuery
		conditionValue := pagination.CValue
		tx := r.db.Scopes(r.paginate(users, &pagination, r.db)).Where(conditionQuery, conditionValue).Find(&users)
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

func (r Userrepo) UpdateUser(user *model.User) error {
	tx := r.db.Save(user)
	return tx.Error
}

func (r Userrepo) DeleteUser(id string) error {
	user, err := r.GetUserById(id)

	if err != nil {
		return err
	}

	tx := r.db.Delete(user)
	return tx.Error
}

func (r Userrepo) CreateUser(user *model.User) (*model.User, error) {
	if err := r.db.Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return user, nil
}

// GetUserById returns user based on given id
func (r Userrepo) GetUserById(id string) (*model.User, error) {
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

// GetUserByNickname returns user based on given username
func (r Userrepo) GetUserByNickname(nickName string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("nick_name = ?", nickName).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail returns user based on given id
func (r Userrepo) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func NewUserRepo(db *gorm.DB, log *log.Logger) UserRepository {
	return &Userrepo{
		db:  db,
		log: log,
	}
}