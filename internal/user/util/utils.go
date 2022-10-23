package util

import (
	"github.com/cemayan/faceit-technical-test/internal/user/model"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"os"
)

func MigrateDB(db *gorm.DB, log *log.Entry) {
	if os.Getenv("ENV") == "test" {
		// ConnectDBForTesting  serves to connect to db for Testing
		// When DB connection is successful then model migration is started
		isExist := db.Migrator().HasTable(&model.User{})
		if isExist {
			err := db.Migrator().DropTable(&model.User{})
			if err != nil {
				return
			}
		}
		err := db.AutoMigrate(&model.User{})
		if err != nil {
			return
		}
		log.Infoln("Database Migrated")
	} else {
		err := db.AutoMigrate(&model.User{})
		if err != nil {
			return
		}
		log.Infoln("Database Migrated")
	}
}
