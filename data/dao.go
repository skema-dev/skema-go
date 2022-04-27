package data

import (
	"reflect"
	"sync"

	"github.com/skema-dev/skema-go/elastic"
	"github.com/skema-dev/skema-go/event"
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

const (
	eventOnDaoCreate = "dao_event_create"
	eventOnDaoUpdate = "dao_event_create"
	eventOnDaoDelete = "dao_event_create"
)

type eventData struct {
	TX    *gorm.DB
	Value DaoModel
}
type DAO struct {
	db            *Database
	model         DaoModel
	es            elastic.Elastic
	columnToField map[string]string
	fieldToColumn map[string]string

	pubsub *event.PubSub
}

func NewDAO(db *Database, model DaoModel) *DAO {
	dao := &DAO{
		db:            db,
		model:         model,
		columnToField: map[string]string{},
		fieldToColumn: map[string]string{},
		pubsub:        event.NewPubSub(),
	}
	dao.initColumnFieldTable()

	// initiate even calling
	f := func(v interface{}) {
		data := v.(*eventData)
		if data.TX.Error != nil {
			logging.Errorf(data.TX.Error.Error())
			return
		}
		if data.TX.RowsAffected == 0 {
			logging.Errorf("nothing changed. no update")
			return
		}

		dao.UpdateElasticIndex(data.Value)
	}

	dao.pubsub.Subscribe(eventOnDaoUpdate, f)
	dao.pubsub.Subscribe(eventOnDaoCreate, f)

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
	defer d.pubsub.Publish(eventOnDaoCreate, &eventData{tx, value})

	if tx.Error != nil {
		logging.Errorf(tx.Error.Error())
	}

	return tx.Error
}

func (d *DAO) Update(query *QueryParams, value DaoModel) error {
	tx := d.db.Where(*query).Updates(value)
	defer d.pubsub.Publish(eventOnDaoCreate, &eventData{tx, value})

	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

// Update if exists (by queryColumns), insert new one if not existing
func (d *DAO) Upsert(value DaoModel, queryColumns []string, assignedColums []string) error {
	var tx *gorm.DB
	defer func() { d.pubsub.Publish(eventOnDaoCreate, &eventData{tx, value}) }()

	if queryColumns == nil || len(queryColumns) == 0 {
		// no query columns exists, jut create new record
		tx = d.db.Create(value)
		return tx.Error
	}

	queries := []clause.Column{}
	for _, col := range queryColumns {
		queries = append(queries, clause.Column{Name: col})
	}

	if assignedColums == nil && len(assignedColums) == 0 {
		// no specific assignment column found, update all
		tx = d.db.Clauses(clause.OnConflict{
			Columns:   queries,
			UpdateAll: true,
		}).Create(value)
		return tx.Error
	}

	// update only assigned column when conflict happends
	tx = d.db.Clauses(clause.OnConflict{
		Columns:   queries,
		DoUpdates: clause.AssignmentColumns(assignedColums),
	}).Create(value)

	return tx.Error
}

func (d *DAO) Query(
	query *QueryParams,
	result interface{},
	options ...QueryOption,
) error {
	if d.es != nil {
		if esSearchErr := d.searchFromElastic(query, result, options...); esSearchErr == nil {
			return nil
		}
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

func (d *DAO) Delete(query interface{}, args ...interface{}) error {
	ch := make(chan error)
	ids := make([]string, 0)

	go func() {
		rs := []map[string]interface{}{}
		tx := d.db.Model(&d.model).Where(query, args...).Find(&rs)
		if tx.Error != nil {
			ch <- logging.Errorf(tx.Error.Error())
			return
		}

		if len(rs) == 0 {
			ch <- logging.Errorf("no matching record found")
			return
		}
		for _, r := range rs {
			ids = append(ids, r["uuid"].(string))
		}
		ch <- nil
	}()
	err := <-ch
	if err != nil {
		return err
	}

	d.DeleteFromElastic(ids)
	tx := d.db.Model(&d.model).Where(query, args...).Delete(&d.model)
	return tx.Error
}

func (d *DAO) esIndexName() string {
	return d.GetDB().Name() + "_" + d.model.TableName()
}

func (d *DAO) UpdateElasticIndex(data DaoModel) {
	if d.es == nil {
		return
	}

	ch := make(chan error)
	go func(c chan error) {
		c <- d.es.Index(d.esIndexName(), data.PrimaryID(), data)
	}(ch)

	result := <-ch
	if result != nil {
		logging.Errorf("update index failed: %s", result.Error())
	}
}

func (d *DAO) DeleteFromElastic(ids []string) {
	if d.es == nil {
		return
	}
	d.es.Delete(d.esIndexName(), ids)
}

func (d *DAO) searchFromElastic(
	query *QueryParams,
	result interface{},
	options ...QueryOption) error {
	if d.es == nil {
		return nil
	}

	newQuery := map[string]interface{}{}
	for k, v := range *query {
		fieldname, ok := d.columnToField[k]
		if !ok {
			logging.Errorf("missing column name: %s", k)
			continue
		}
		newQuery[fieldname] = v
	}

	var searchOption *elastic.SearchOption

	if len(options) > 0 {
		searchOption = &elastic.SearchOption{
			From: options[0].Offset,
			Size: options[0].Limit,
			Sort: options[0].Order,
		}
	}

	founds, err := d.es.Search(d.esIndexName(), "match", newQuery, searchOption)
	if err != nil {
		return logging.Errorf("Error happend when search from elastic for %s: %s", d.esIndexName(), err.Error())
	}

	modelType := reflect.TypeOf(d.model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

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
