package postgres

import (
	"fmt"
	"github.com/cemayan/faceit-technical-test/pkg/common"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type DBHandler interface {
	New() *gorm.DB
}

type DBService struct {
	configs *common.Postgresql
}

// New  serves to connect to db
// When DB connection is successful then model migration is started
func (d DBService) New() *gorm.DB {

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color
		},
	)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: fmt.Sprintf("host=%s port=%s  user=%s password=%s  dbname=%s sslmode=disable ",
			d.configs.HOST,
			d.configs.PORT,
			d.configs.USER,
			d.configs.PASSWORD,
			d.configs.NAME),
	}), &gorm.Config{Logger: newLogger})

	if err != nil {
		panic("failed to connect database")
	}

	fmt.Println("Connection Opened to Database")

	return db
}

func NewDbHandler(configs *common.Postgresql) DBHandler {
	return &DBService{configs: configs}
}