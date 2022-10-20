package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// DB gorm connector
var DB *gorm.DB
var MongoDBClient *mongo.Client
var MongoDBContext = context.TODO()
