package data

import (
	"errors"
	"reflect"
	"sync"

	"github.com/skema-dev/skema-go/elastic"
	"github.com/skema-dev/skema-go/logging"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Shortcut for map based query parameters defined in gorm
type QueryParams = map[string]interface{}

// Options for ordery by, limit and offset. This is NOT required
type QueryOption struct {
	Order  string
	Limit  int
	Offset int
}

type DaoModel interface {
	TableName() string
	PrimaryID() string
}

type DAO struct {
	db            *Database
	model         DaoModel
	es            elastic.Elastic
	columnToField map[string]string
	fieldToColumn map[string]string
}

func NewDAO(db *Database, model DaoModel) *DAO {
	dao := &DAO{
		db:            db,
		model:         model,
		columnToField: map[string]string{},
		fieldToColumn: map[string]string{},
	}
	dao.initColumnFieldTable()

	modelValue := reflect.ValueOf(dao.model)
	params := []reflect.Value{reflect.ValueOf(dao)}
	method := modelValue.MethodByName("SetDAO")
	if !method.IsValid() {
		logging.Fatalf("incorrect model type. Make sure your model contains data.Model field")
	}
	method.Call(params)

	return dao
}

func toPtr(v reflect.Value) reflect.Value {
	pt := reflect.PtrTo(v.Type()) // create a *T type.
	pv := reflect.New(pt.Elem())  // create a reflect.Value of type *T.
	pv.Elem().Set(v)              // sets pv to point to underlying value of v.
	return pv
}

// return the raw gorm.DB when necessary, so user can perform whatever the standard gorm can do.
func (d *DAO) GetDB() *Database {
	return d.db
}

func (d *DAO) Name() string {
	return d.model.TableName()
}

func (d *DAO) SetElasticClient(client elastic.Elastic) {
	d.es = client
}

func (d *DAO) Automigrate() {
	d.db.AutoMigrate(d.model)
}

func (d *DAO) Create(value DaoModel) error {
	tx := d.db.Create(value)
	if tx.Error != nil {
		logging.Errorf(tx.Error.Error())
	}
	return tx.Error
}

func (d *DAO) Update(query *QueryParams, value DaoModel) error {
	tx := d.db.Model(&d.model).Where(*query).Updates(value)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("Doesn't found matching row to update")
	}

	return nil
}

// Update if exists (by queryColumns), insert new one if not existing
func (d *DAO) Upsert(value DaoModel, queryColumns []string, assignedColums []string) error {
	var result *gorm.DB

	if queryColumns == nil || len(queryColumns) == 0 {
		// no query columns exists, jut create new record
		result = d.db.Create(value)
		return result.Error
	}

	queries := []clause.Column{}
	for _, col := range queryColumns {
		queries = append(queries, clause.Column{Name: col})
	}

	if assignedColums == nil && len(assignedColums) == 0 {
		// no specific assignment column found, update all
		result = d.db.Clauses(clause.OnConflict{
			Columns:   queries,
			UpdateAll: true,
		}).Create(value)
		return result.Error
	}

	// update only assigned column when conflict happends
	result = d.db.Clauses(clause.OnConflict{
		Columns:   queries,
		DoUpdates: clause.AssignmentColumns(assignedColums),
	}).Create(value)
	return result.Error
}

func (d *DAO) Query(
	query *QueryParams,
	result interface{},
	options ...QueryOption,
) error {
	if d.es != nil {
		return d.searchFromElastic(query, result, options...)
	}

	var tx *gorm.DB
	if len(options) == 0 {
		tx = d.db.Model(&d.model).Where(*query).Find(result)
	} else {
		option := options[0]
		tx = d.db.Model(&d.model).Where(*query)
		if len(option.Order) > 0 {
			tx = tx.Order(option.Order)
		}
		if option.Offset > 0 {
			tx = tx.Offset(option.Offset)
		}
		if option.Limit > 0 {
			tx = tx.Limit(option.Limit)
		}
		tx.Find(result)
	}

	if tx.Error != nil {
		logging.Errorf(
			"query failed for [%s]. %v :  %s",
			d.model.TableName(),
			*query,
			tx.Error.Error(),
		)
	}

	return tx.Error
}

func (d *DAO) Delete(conds ...interface{}) error {
	tx := d.db.Delete(&d.model, conds...)
	return tx.Error
}

func (d *DAO) esIndexName() string {
	return d.GetDB().Name() + "_" + d.model.TableName()
}

func (d *DAO) UpdateElasticIndex(data interface{}) {
	if d.es == nil {
		return
	}

	d.es.Index(d.esIndexName(), d.model.PrimaryID(), data)
}

func (d *DAO) DeleteFromElastic(id string) {
	if d.es != nil {
		d.es.Delete(d.esIndexName(), id)
	}
}

func (d *DAO) searchFromElastic(
	query *QueryParams,
	result interface{},
	options ...QueryOption) error {
	if d.es == nil {
		return nil
	}

	newQuery := QueryParams{}
	for k, v := range *query {
		fieldname, ok := d.columnToField[k]
		if !ok {
			logging.Errorf("missing column name: %s", k)
			continue
		}
		newQuery[fieldname] = v
	}

	founds, err := d.es.Search(d.esIndexName(), "match", newQuery)
	if err != nil {
		return logging.Errorf("Error happend when search from elastic for %s: %s", d.esIndexName(), err.Error())
	}

	modelType := reflect.TypeOf(d.model)

	reflectedResult := reflect.ValueOf(result)
	if reflect.TypeOf(result).Kind() == reflect.Ptr {
		reflectedResult = reflectedResult.Elem()
	}

	items := reflect.MakeSlice(reflect.SliceOf(modelType), 0, len(founds))

	for _, found := range founds {
		value := reflect.New(modelType)
		data := value.Interface()
		elastic.ConvertMapToStruct(found, data)
		dataReflectValue := reflect.ValueOf(data).Elem()
		items = reflect.Append(items, dataReflectValue)
	}

	convertedValue := items.Convert(reflectedResult.Type())
	reflectedResult.Set(convertedValue)

	return nil
}

func (d *DAO) initColumnFieldTable() {
	s, err := schema.Parse(&d.model, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		panic("failed to create schema")
	}

	for _, field := range s.Fields {
		dbName := field.DBName
		modelName := field.Name
		d.columnToField[dbName] = modelName
		d.fieldToColumn[modelName] = dbName
	}
}
