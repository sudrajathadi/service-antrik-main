package models

import (
	"time"
)

type AppointmentStatus string

const (
	StatusPending   AppointmentStatus = "pending"
	StatusConfirmed AppointmentStatus = "confirmed"
	StatusCancelled AppointmentStatus = "cancelled"
	StatusDone      AppointmentStatus = "done"
)

type Appointment struct {
	Base
	UserID          uint              `json:"user_id"          gorm:"not null;index"`
	DoctorID        uint              `json:"doctor_id"        gorm:"not null;index"`
	HospitalID      uint              `json:"hospital_id"      gorm:"not null;index"`
	AppointmentDate time.Time         `json:"appointment_date" gorm:"not null"`
	AppointmentTime string            `json:"appointment_time" gorm:"type:varchar(5);not null"` // e.g. "09:00"
	SymptomsNote    string            `json:"symptoms_note"    gorm:"type:text"`
	Status          AppointmentStatus `json:"status"           gorm:"type:varchar(20);default:'pending'"`
	User            User              `json:"user"             gorm:"foreignKey:UserID"`
	Doctor          Doctor            `json:"doctor"           gorm:"foreignKey:DoctorID"`
	Hospital        Hospital          `json:"hospital"         gorm:"foreignKey:HospitalID"`
}
