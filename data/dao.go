package data

import (
	"fmt"

	"github.com/skema-dev/skema-go/logging"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
}

type DAO struct {
	db    *Database
	model DaoModel
}

func NewDAO(db *Database, model DaoModel) *DAO {
	return &DAO{db: db, model: model}
}

// return the raw gorm.DB when necessary, so user can perform whatever the standard gorm can do.
func (d *DAO) GetDB() *Database {
	return d.db
}

func (d *DAO) Name() string {
	return d.model.TableName()
}

func (d *DAO) Automigrate() {
	d.db.AutoMigrate(d.model)
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
		fmt.Println("update all")
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
	var tx *gorm.DB

	if len(options) == 0 {
		tx = d.db.Model(d.model).Where(*query).Find(result)
	} else {
		option := options[0]
		tx = d.db.Model(d.model).Where(*query)
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
	tx := d.db.Delete(d.model, conds...)
	return tx.Error
}

func (d *DAO) BatchDelete(condition string) error {
	tx := d.db.Where(condition).Delete(d.model)
	if tx.Error != nil {
		logging.Errorw(tx.Error.Error(), "condition", condition, "modelname", d.model.TableName())
	}
	return tx.Error
}
