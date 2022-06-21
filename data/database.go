package data

import (
	"errors"
	"fmt"
	"time"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/elastic"
	"github.com/skema-dev/skema-go/logging"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// wrap standard gorm.DB. For now, it's not doing much.
type Database struct {
	gorm.DB
	automigrate   bool
	elasticClient elastic.Elastic
}

func (d Database) ShouldAutomigrate() bool {
	return d.automigrate
}

func (d *Database) SetElastic(client elastic.Elastic) {
	d.elasticClient = client
}

func (d *Database) Elastic() elastic.Elastic {
	return d.elasticClient
}

// initiate mysql db and return the instance
func NewMysqlDatabase(conf *config.Config) (*Database, error) {
	username := conf.GetString("username")
	password := conf.GetString("password")
	host := conf.GetString("host")
	port := conf.GetInt("port")
	dbname := conf.GetString("dbname")
	charset := conf.GetString("charset", "utf8mb4")
	retries := conf.GetInt("retry", 0)

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		username,
		password,
		host,
		port,
		dbname,
		charset,
	)

	logging.Debugf("connecting to %s", dsn)
	db := retryConnectDatabase(retries, func() (*gorm.DB, error) {
		return gorm.Open(mysql.Open(dsn), &gorm.Config{})
	})
	if db == nil {
		return nil, errors.New("failed to connect db")
	}

	logging.Debugf("connectedto %s", dsn)

	return &Database{
		DB:          *db,
		automigrate: conf.GetBool("automigrate", false),
	}, nil
}

// initiate a sqlite db  for in-memeory implementing, and return the instance
func NewMemoryDatabase(conf *config.Config) (*Database, error) {
	db := retryConnectDatabase(0, func() (*gorm.DB, error) {
		return gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	})
	if db == nil {
		return nil, errors.New("failed to connect db")
	}

	return &Database{
		DB:          *db,
		automigrate: true,
	}, nil
}

// initiate sqlite db and return the instance
func NewSqliteDatabase(conf *config.Config) (*Database, error) {
	dbfile := conf.GetString("filepath")
	if dbfile == "" {
		return nil, errors.New("sqlite filepath is not defined")
	}

	db := retryConnectDatabase(0, func() (*gorm.DB, error) {
		return gorm.Open(sqlite.Open(dbfile), &gorm.Config{})
	})
	if db == nil {
		return nil, errors.New("failed to connect db")
	}

	return &Database{
		DB:          *db,
		automigrate: conf.GetBool("automigrate", false),
	}, nil
}

// initiate postgresql db and return the instance
func NewPostsqlDatabase(conf *config.Config) (*Database, error) {
	username := conf.GetString("username")
	password := conf.GetString("password")
	host := conf.GetString("host")
	port := conf.GetInt("port")
	dbname := conf.GetString("dbname")
	timezone := conf.GetString("timezone", "Asia/Shanghai")
	retries := conf.GetInt("retry", 0)

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=%s",
		host,
		username,
		password,
		dbname,
		port,
		timezone,
	)
	options := conf.GetString("options")
	if len(options) > 0 {
		dsn += "options=" + options
	}

	logging.Debugf("connecting to %s", dsn)

	db := retryConnectDatabase(retries, func() (*gorm.DB, error) {
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	})
	if db == nil {
		return nil, errors.New("failed to connect db")
	}

	logging.Debugf("connectedto %s", dsn)

	return &Database{
		DB:          *db,
		automigrate: conf.GetBool("automigrate", false),
	}, nil
}

func retryConnectDatabase(retryTimes int, fn func() (*gorm.DB, error)) *gorm.DB {
	i := 0
	for i <= retryTimes {
		db, err := fn()
		if err == nil {
			return db
		}
		logging.Errorf(err.Error())

		i += 1
		if i <= retryTimes {
			logging.Errorf("Failed to connect db. Attempt to retry in 3 seconds")
			time.Sleep(3 * time.Second)
		}
	}

	return nil
}
