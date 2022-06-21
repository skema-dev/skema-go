package data

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	gorm.Model
	UUID string `gorm:"column:uuid;size:64;not null;uniqueIndex"`
}

func (m Model) PrimaryID() string {
	return m.UUID
}

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if m.UUID == "" {
		m.UUID = uuid.New().String()
	}

	return nil
}

func (m *Model) BeforeSave(tx *gorm.DB) error {
	if m.UUID == "" {
		m.UUID = uuid.New().String()
	}

	return nil
}
