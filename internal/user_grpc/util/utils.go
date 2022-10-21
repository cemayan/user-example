package util

import (
	"fmt"
	"github.com/cemayan/faceit-technical-test/internal/user/model"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"os"
	"time"
)

// FailOnError returns a log based on given error and message
func FailOnError(err error, msg string) {
	if err != nil {
		log.Errorf("%s: %s", msg, err)
	}
}

func GetTime() int64 {
	now := time.Now()
	return now.Unix()
}

func MigrateDB(db *gorm.DB) {
	if os.Getenv("ENV") == "test" {
		// ConnectDBForTesting  serves to connect to db for Testing
		// When DB connection is successful then model migration is started
		db.Migrator().DropTable(&model.User{})
		db.AutoMigrate(&model.User{})
		fmt.Println("Database Migrated")
	} else {
		db.AutoMigrate(&model.User{})
		fmt.Println("Database Migrated")
	}
}
