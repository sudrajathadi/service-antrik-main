package models

type User struct {
	Base
	ChatID      string `json:"chat_id"      gorm:"type:varchar(100);uniqueIndex;not null"`
	FullName    string `json:"full_name"    gorm:"type:varchar(255);not null"`
	PhoneNumber string `json:"phone_number" gorm:"type:varchar(20)"`
	Email       string `json:"email"        gorm:"type:varchar(255);uniqueIndex"`
}
