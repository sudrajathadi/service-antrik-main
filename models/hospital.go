package models

type Hospital struct {
	Base
	Name        string `json:"name"         gorm:"type:varchar(255);not null"`
	Address     string `json:"address"      gorm:"type:text;not null"`
	City        string `json:"city"         gorm:"type:varchar(100);not null"`
	PhoneNumber string `json:"phone_number" gorm:"type:varchar(20)"`
}
