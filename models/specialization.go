package models

type Specialization struct {
	Base
	Name        string `json:"name"        gorm:"type:varchar(255);not null;uniqueIndex"`
	Description string `json:"description" gorm:"type:text"`
}
