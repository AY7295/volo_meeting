package model

type Device struct {
	Id       string    `json:"id" gorm:"type:varchar(20);primary_key"`
	Meetings []Meeting `json:"meetings" gorm:"many2many:meeting_device;"`
}
