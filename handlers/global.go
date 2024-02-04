package handlers

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"gopkg.in/ini.v1"
	"io"
	"log"
	"open-indexer/utils"
	"os"
	"time"
)

var cfg *ini.File
var logger *logrus.Logger
var db *leveldb.DB

var DataSourceType string

var mgCtx *context.Context
var mongodb *mongo.Database

var snapFile string

var QuitChan = make(chan bool)
var StopSuccessCount uint = 0

func init() {
	log.Println("global init")
	var snapshotAt string
	flag.StringVar(&snapFile, "snapshot", "", "the filename of snapshot")
	flag.StringVar(&snapshotAt, "snapshot-at", "", "the block that create snapshot")
	flag.Parse()

	if snapshotAt != "" {
		createSnapshotBlock = uint64(utils.ParseInt64(snapshotAt))
	}

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
	initLevelDb()

	initDataSource()

	initSync()

	// read data
	err = initFromStorage()
	if err != nil {
		panic(err)
	}
}

func initLogger() {
	writerStd := os.Stdout
	writerFile, err := os.OpenFile("logs.txt", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		logrus.Fatalf("create file logs.txt failed: %v", err)
	}

	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{})
	logger.SetOutput(io.MultiWriter(writerStd, writerFile))
}

func initLevelDb() {
	dbCfg := cfg.Section("leveldb")
	dbPath := dbCfg.Key("path").String()

	var err error
	if snapFile != "" {
		_, err := os.Stat(dbPath)
		if err == nil {
			panic("when specifying a snapshot, the database must be empty")
		}
	}

	db, err = leveldb.OpenFile(dbPath, nil)
	if err != nil {
		panic("open database failed:" + err.Error())
	}
}

func initDataSource() {
	dsCfg := cfg.Section("data-source")
	DataSourceType = dsCfg.Key("type").String()
	dsUri := dsCfg.Key("uri").String()

	if DataSourceType == "mongo" {
		cs, err := connstring.ParseAndValidate(dsUri)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		mgCtx = &ctx
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsUri))
		if err != nil {
			panic("connect to mongo failed:" + err.Error())
		}
		mongodb = client.Database(cs.Database)
	} else if DataSourceType == "rpc" {
		fetchUrl = dsUri
	} else {
		panic("error data source type")
	}
}

func GetLogger() *logrus.Logger {
	return logger
}

func CloseDb() {
	db.Close()
	db = nil
	if mongodb != nil {
		mongodb.Client().Disconnect(*mgCtx)
	}
	mongodb = nil
}
