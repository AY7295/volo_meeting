package model

import (
	"gorm.io/gorm"
	"time"
)

type Meeting struct {
	Id         string     `json:"id" gorm:"type:varchar(20);primary_key"`
	FriendlyId string     `json:"friendly_id" gorm:"type:varchar(20);index;not null"`
	StartTime  *time.Time `json:"start_time" gorm:"type:datetime;index;not null"`
	EndTime    *time.Time `json:"end_time" gorm:"type:datetime;index;not null"`

	Devices []Device `json:"devices" gorm:"many2many:meeting_device;"`
}

func (m *Meeting) Create(db *gorm.DB) error {
	return db.Model(m).Create(m).Error
}

func (m *Meeting) FindById(db *gorm.DB) error {
	return db.Model(m).Where("id = ?", m.Id).First(m).Error
}

func (m *Meeting) AppendDevice(db *gorm.DB, device *Device) error {
	return db.Model(m).Association("Devices").Append(&device)
}

func (m *Meeting) Update(db *gorm.DB, updates map[string]any) error {
	return db.Model(m).Updates(updates).Error
}
