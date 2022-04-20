package data

import (
	"strings"
	"sync"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
)

// DataManager provides simple interface to loop up a db instance/dao/etc.
type DataManager struct {
	databases map[string]*Database
	daoMap    sync.Map
}

var (
	dbCreateMap = map[string]func(*config.Config) (*Database, error){
		"mysql":  NewMysqlDatabase,
		"memory": NewMemoryDatabase,
		"sqlite": NewSqliteDatabase,
		"pgsql":  NewPostsqlDatabase,
	}

	dataMan *DataManager
)

func InitWithConfigFile(filepath string, key string) {
	conf := config.NewConfigWithFile(filepath)
	InitWithConfig(conf, key)
}

func InitWithConfig(conf *config.Config, key string) {
	dataMan = NewDataManager().WithConfig(conf, key)
}

func Manager() *DataManager {
	return dataMan
}

func NewDataManager() *DataManager {
	man := &DataManager{
		databases: map[string]*Database{},
		daoMap:    sync.Map{},
	}
	return man
}

func (d *DataManager) WithConfig(conf *config.Config, key string) *DataManager {
	if conf == nil {
		return d
	}

	confs := conf.GetMapConfig(key)
	for k, v := range confs {
		d.AddDatabaseWithConfig(&v, k)
	}

	return d
}

func (d *DataManager) AddDatabaseWithConfig(conf *config.Config, dbKey string) {
	if dbKey == "" {
		logging.Fatalf("AddDatabaseWithConfig must specify a key for the db!")
	}

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

func (d DataManager) GetDB(dbKey string) *Database {
	if dbKey == "" {
		// no key specified, return the db if there is only one, otherwise fatal exit
		if len(d.databases) > 1 {
			logging.Fatalf("more than 1 database defined. Please specify the exact db with a key")
		}

		for _, v := range d.databases {
			logging.Debugf("no database key specified, return the default db")
			return v
		}
	}

	db, ok := d.databases[dbKey]
	if !ok {
		logging.Errorf("cannot find database with key %s", dbKey)
		return nil
	}

	return db
}

// register dao models for the default database
func (d *DataManager) RegisterDaoModels(models []DaoModel) {
	d.RegisterDaoModelsForDb("", models)
}

// register dao models for the specified database
func (d *DataManager) RegisterDaoModelsForDb(dbKey string, models []DaoModel) {
	for _, model := range models {
		d.RegisterDaoForDb(dbKey, model)
	}
}

func (d *DataManager) RegisterDao(model DaoModel) {
	d.RegisterDaoForDb("", model)
}

// register A dao model for the specified database
func (d *DataManager) RegisterDaoForDb(dbKey string, model DaoModel) {
	db := d.GetDB(dbKey)
	if db == nil {
		logging.Fatalf("incorrect dbKey when init dao in db manager: %s", dbKey)
	}

	daoKey := d.getDaoKey(dbKey, model)
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

func (d DataManager) GetDAO(model DaoModel) *DAO {
	return d.GetDaoForDb("", model)
}

func (d DataManager) GetDaoForDb(dbKey string, model DaoModel) *DAO {
	daoKey := d.getDaoKey(dbKey, model)
	v, ok := d.daoMap.Load(daoKey)
	if !ok {
		logging.Fatalf("dao not initialized for %s", daoKey)
	}
	return v.(*DAO)
}

func (d DataManager) getDaoKey(dbKey string, model DaoModel) string {
	key := "default"
	if dbKey != "" {
		key = dbKey
	}

	return key + "." + model.TableName()
}
