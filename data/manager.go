package data

import (
	"reflect"
	"strings"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/elastic"
	"github.com/skema-dev/skema-go/logging"
)

// DataManager provides simple interface to loop up a db instance/dao/etc.
type DataManager struct {
	databases map[string]*Database
	// [db_key:[table_name:model]]
	daoMap map[string]map[string]DAO

	elasticClient elastic.Elastic
}

var (
	dbCreateMap = map[string]func(*config.Config) (*Database, error){
		"mysql":  NewMysqlDatabase,
		"memory": NewMemoryDatabase,
		"sqlite": NewSqliteDatabase,
		"pgsql":  NewPostsqlDatabase,
	}

	// model type registry: [package:[typeName: type]]
	modelTypeRegistry = make(map[string]map[string]reflect.Type)

	dataMan *DataManager
)

func InitWithFile(filepath string, key string) {
	conf := config.NewConfigWithFile(filepath)
	InitWithConfig(conf, key)
}

func InitWithConfig(conf *config.Config, key string) {
	dataMan = NewDataManager().WithConfig(conf, key)
}

func R(model DaoModel) {
	t := reflect.TypeOf(model).Elem()
	pkgPath := t.PkgPath()
	typeName := t.Name()

	modelTypes, ok := modelTypeRegistry[t.PkgPath()]
	if !ok {
		modelTypes = make(map[string]reflect.Type)
		modelTypeRegistry[pkgPath] = modelTypes
	}
	if _, ok := modelTypes[typeName]; ok {
		logging.Warnw("model type already exists and will be overwritten.", "package", pkgPath, "type", typeName)
	}

	modelTypes[typeName] = t

	logging.Debugw("DaoModel Type registered", "package", pkgPath, "type", typeName)
}

func Manager() *DataManager {
	return dataMan
}

func NewDataManager() *DataManager {
	man := &DataManager{
		databases: map[string]*Database{},
		daoMap:    map[string]map[string]DAO{},
	}
	return man
}

func (d *DataManager) WithConfig(conf *config.Config, key string) *DataManager {
	if conf == nil {
		return d
	}

	confs := conf.GetMapConfig(key)
	for k, v := range confs {
		d.AddDatabaseWithConfig(&v, k, conf)
	}

	return d
}

func (d *DataManager) AddDatabaseWithConfig(conf *config.Config, dbKey string, originalConfig *config.Config) {
	logging.Debugf("Add Database for %s", dbKey)
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

	models := conf.GetMapFromArray("models")
	if models != nil {
		d.initDaoModelForDb(dbKey, models)
	}

	// check if elasticsearch is defined
	queryConf := conf.GetSubConfig("cqrs")
	if queryConf != nil {
		queryType := queryConf.GetString("type")
		if queryType == "elastic" {
			elasticConfigKey := queryConf.GetString("name")
			client := elastic.NewElasticClient(originalConfig.GetSubConfig(elasticConfigKey))
			db.SetElastic(client)
		}
	}

}

//
//
//     name1: //no package specified, look through all type registry maps
//     name2:
//        package: xxxxxx (optional)
//
//
func (d *DataManager) initDaoModelForDb(dbkey string, models map[string]interface{}) {
	for modelTypeName, v := range models {
		var daoModel DaoModel

		if v == nil {
			// no package specified, look into every registry map (in most cases, it's just one map)
			logging.Debugf("no package specified")
			daoModel = d.findModelType(modelTypeName)
		} else {
			confMap := v.(map[interface{}]interface{})

			if pkg, ok := confMap["package"]; ok {
				// package specified, look into the specific registry map
				types, ok := modelTypeRegistry[pkg.(string)]
				if !ok {
					logging.Fatalw("incorrect package when initializing dao", "package", pkg, "mode type", modelTypeName)
				}
				modelType, ok := types[modelTypeName]
				if !ok {
					logging.Fatalw("incorrect type name when initializing dao", "package", pkg, "mode type", modelTypeName, "types", types)
				}
				daoModel = reflect.New(modelType).Elem().Interface().(DaoModel)
			} else {
				daoModel = d.findModelType(modelTypeName)
			}
		}

		if daoModel == nil {
			logging.Fatalw("incorrect definition for model", "model name", modelTypeName, "config", v)
		}

		d.GetDaoForDb(dbkey, daoModel)

		db := d.GetDB(dbkey)

		if db.automigrate {
			db.AutoMigrate(daoModel)
		}
	}
}

// find model type in the whole type registry tables
func (d DataManager) findModelType(modelTypeName string) DaoModel {
	for _, models := range modelTypeRegistry {
		modelType, ok := models[modelTypeName]
		if ok {
			return reflect.New(modelType).Elem().Interface().(DaoModel)
		}
	}
	return nil
}

// Get the underlying database object
func (d DataManager) GetDB(dbKey string) *Database {
	if dbKey == "" {
		// no key specified, return the db if there is only one, otherwise fatal exit
		if len(d.databases) > 1 {
			logging.Errorf("more than 1 database defined. Please specify the exact db with a key")
			return nil
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

func (d *DataManager) GetDAO(model DaoModel) *DAO {
	return d.GetDaoForDb("", model)
}

// register A dao model for the specified database
func (d *DataManager) GetDaoForDb(dbKey string, model DaoModel) *DAO {
	db := d.GetDB(dbKey)
	if db == nil {
		logging.Errorf("incorrect dbKey when init dao in db manager: %s", dbKey)
		return nil
	}

	dbs, ok := d.daoMap[dbKey]
	if !ok {
		dbs = make(map[string]DAO)
		d.daoMap[dbKey] = dbs
	}

	daoIns, ok := d.daoMap[dbKey][model.TableName()]
	if ok {
		return &daoIns
	}

	newDao := NewDAO(db, model)
	dbs[model.TableName()] = *newDao
	logging.Debugw("DAO not found. New DAO created", "db", dbKey, "table", model.TableName())

	newDao.SetElasticClient(db.Elastic())
	// now initialize the table if necessary
	if db.ShouldAutomigrate() {
		db.AutoMigrate(model)
	}

	return newDao
}
