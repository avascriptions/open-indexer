package handlers

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"gopkg.in/ini.v1"
	"io"
	"os"
	"time"
)

var cfg *ini.File
var logger *logrus.Logger
var mongodb *mongo.Database

func init() {

	var err error
	cfg, err = ini.ShadowLoad("config.ini")
	if err != nil {
		cfg, err = ini.ShadowLoad("../config.ini")
		if err != nil {
			panic("read config.ini file error: " + err.Error())
			//os.Exit(-1)
		}
	}

	initLogger()
	initMongo()
}

func initLogger() {
	writerStd := os.Stdout
	writerFile, err := os.OpenFile("logs.txt", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		logrus.Fatalf("create file log.txt failed: %v", err)
	}

	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{})
	logger.SetOutput(io.MultiWriter(writerStd, writerFile))
}

func initMongo() {

	dbCfg := cfg.Section("mongo")
	dbUri := dbCfg.Key("uri").String()

	cs, err := connstring.ParseAndValidate(dbUri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbUri))
	if err != nil {
		logrus.Fatalf("connect to mongo failed: %v", err)
	}
	mongodb = client.Database(cs.Database)
}

func GetLogger() *logrus.Logger {
	return logger
}
