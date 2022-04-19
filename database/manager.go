package database

import (
	"reflect"
	"strings"
	"sync"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
)

// DatabaseManager provides simple interface to loop up a db instance
type DatabaseManager struct {
	databases map[string]*Database
	daoMap    sync.Map
	daoTypes  map[string]reflect.Type
}

var (
	dbCreateMap = map[string]func(*config.Config) (*Database, error){
		"mysql":  NewMysqlDatabase,
		"memory": NewMemoryDatabase,
		"sqlite": NewSqliteDatabase,
		"pgsql":  NewPostsqlDatabase,
	}

	dbMan *DatabaseManager
)

func InitWithConfig(conf *config.Config, key string) {
	dbMan = NewDatabaseManager().WithConfig(conf, key)
}

func Manager() *DatabaseManager {
	return dbMan
}

func NewDatabaseManager() *DatabaseManager {
	man := &DatabaseManager{
		databases: map[string]*Database{},
		daoMap:    sync.Map{},
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

func (d *DatabaseManager) AddDatabaseWithConfig(conf *config.Config, dbKey string) {
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

	d.databases[dbKey] = db
}

func (d DatabaseManager) GetDB(dbKey string) *Database {
	db, ok := d.databases[dbKey]
	if !ok {
		logging.Errorf("cannot find database with key %s", dbKey)
		return nil
	}

	return db
}

func (d *DatabaseManager) RegisterDaoModels(dbKey string, models []DaoModel) {
	for _, model := range models {
		d.RegisterDAO(dbKey, model)
	}
}

func (d *DatabaseManager) RegisterDAO(dbKey string, model DaoModel) {
	db := d.GetDB(dbKey)
	if db == nil {
		logging.Fatalf("incorrect dbKey when init dao in db manager: %s", dbKey)
	}

	daoKey := d.GetDaoKey(dbKey, model)
	newDao := &DAO{db: db, model: model}

	if _, loaded := d.daoMap.LoadOrStore(daoKey, newDao); loaded {
		// dao already created, just return the existing one
		logging.Debugf("dao alreay exists: %s", daoKey)
		return
	}

	// now initialize the table if necessary
	if db.ShouldAutomigrate() {
		db.AutoMigrate(model)
	}
}

func (d DatabaseManager) GetDAO(dbKey string, model DaoModel) *DAO {
	daoKey := d.GetDaoKey(dbKey, model)
	v, ok := d.daoMap.Load(daoKey)
	if !ok {
		logging.Fatalf("dao not initialized for %s", daoKey)
	}
	return v.(*DAO)
}

func (d DatabaseManager) GetDaoKey(dbKey string, model DaoModel) string {
	return dbKey + "." + model.TableName()
}
