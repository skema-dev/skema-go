package data

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	gorm.Model
	UUID string `gorm:"column:uuid;not null;uniqueIndex"`
	dao  *DAO   `gorm:"-:all"  json:"-"`
}

func (m Model) PrimaryID() string {
	return m.UUID
}

func (m *Model) SetDAO(dao *DAO) {
	m.dao = dao
}

func (m *Model) BeforeCreate(tx *gorm.DB) (err error) {
	if m.UUID == "" {
		m.UUID = uuid.New().String()
	}
	return
}

func (m *Model) BeforeSave(tx *gorm.DB) (err error) {
	if err = m.BeforeCreate(tx); err != nil {
		return err
	}

	return nil
}

func (m *Model) AfterCreate(tx *gorm.DB) (err error) {
	if m.dao != nil {
		m.dao.UpdateElasticIndex(m)
	}

	return
}

func (m *Model) AfterSave(tx *gorm.DB) (err error) {
	fmt.Printf("model.dao: %v\n", m.dao)
	if m.dao != nil {
		m.dao.UpdateElasticIndex(m)
	}

	return
}

func (m *Model) AfterUpdate(tx *gorm.DB) (err error) {
	fmt.Printf("model.dao: %v\n", m.dao)
	if m.dao != nil {
		m.dao.UpdateElasticIndex(m)
	}

	return
}

func (m *Model) AfterDelete(tx *gorm.DB) (err error) {
	if m.dao != nil {
		m.dao.DeleteFromElastic(m.PrimaryID())
	}

	return
}
