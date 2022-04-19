package database

import (
	"errors"
	"fmt"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// wrap standard gorm.DB. For now, it's not doing much.
type Database struct {
	gorm.DB
	automigrate bool
}

func (d Database) ShouldAutomigrate() bool {
	return d.automigrate
}

// initiate mysql db and return the instance
func NewMysqlDatabase(conf *config.Config) (*Database, error) {
	username := conf.GetString("username")
	password := conf.GetString("password")
	host := conf.GetString("host")
	port := conf.GetInt("port")
	dbname := conf.GetString("dbname")
	charset := conf.GetString("charset", "utf8mb4")
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
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logging.Errorf(err.Error())
		return nil, err
	}
	logging.Debugf("connectedto %s", dsn)

	return &Database{
		DB:          *db,
		automigrate: conf.GetBool("automigrate", true),
	}, nil
}

// initiate a sqlite db  for in-memeory implementing, and return the instance
func NewMemoryDatabase(conf *config.Config) (*Database, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, err
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

	db, err := gorm.Open(sqlite.Open(dbfile), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	fmt.Printf("sqlite automigrate: %t\n", conf.GetBool("automigrate"))
	return &Database{
		DB:          *db,
		automigrate: conf.GetBool("automigrate", true),
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
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logging.Errorf(err.Error())
		return nil, err
	}
	logging.Debugf("connectedto %s", dsn)

	return &Database{
		DB:          *db,
		automigrate: conf.GetBool("automigrate", true),
	}, nil
}
