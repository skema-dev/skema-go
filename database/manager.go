package database

import (
	"strings"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
)

// DatabaseManager provides simple interface to loop up a db instance
type DatabaseManager struct {
	databases map[string]*Database
}

var (
	dbCreateMap = map[string]func(*config.Config) (*Database, error){
		"mysql":  NewMysqlDatabase,
		"memory": NewMemoryDatabase,
		"sqlite": NewSqliteDatabase,
		"pgsql":  NewPostsqlDatabase,
	}
)

func NewDatabaseManager() *DatabaseManager {
	man := &DatabaseManager{
		databases: map[string]*Database{},
	}
	return man
}

func (d *DatabaseManager) WithConfig(conf *config.Config, key string) *DatabaseManager {
	if conf == nil {
		return d
	}

	confs := conf.GetMapConfig(key)
	for k, v := range confs {
		d.AddDatabaseWithConfig(&v, k)
	}

	return d
}

func (d *DatabaseManager) AddDatabaseWithConfig(conf *config.Config, opts ...string) {
	dbtype := conf.GetString("type")
	dbtype = strings.ToLower(dbtype)
	createFn, ok := dbCreateMap[dbtype]
	if !ok {
		logging.Fatalf("database type %s is not supported", dbtype)
	}

	db, err := createFn(conf)
	if err != nil {
		logging.Fatalf("failed creating database")
	}

	dbKey := "default"
	if len(opts) > 0 {
		dbKey = opts[0]
	}
	d.databases[dbKey] = db
}

func (d *DatabaseManager) GetDB(dbKey string) *Database {
	db, ok := d.databases[dbKey]
	if !ok {
		logging.Errorf("cannot find database with key %s", dbKey)
		return nil
	}

	return db
}
