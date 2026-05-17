package models

type DoctorSchedule struct {
	Base
	DoctorID     uint   `json:"doctor_id"     gorm:"not null;index"`
	DayOfWeek    string `json:"day_of_week"   gorm:"type:varchar(10);not null"` // "Monday", "Tuesday", etc.
	StartTime    string `json:"start_time"    gorm:"type:varchar(5);not null"`  // Changed to varchar(5) for "HH:MM" consistency
	EndTime      string `json:"end_time"      gorm:"type:varchar(5);not null"`  // e.g., "12:00", "15:00"
	SlotInterval int    `json:"slot_interval" gorm:"default:30"`                // 5, 10, 15, 30, 60 etc.
	Doctor       Doctor `json:"doctor"        gorm:"foreignKey:DoctorID"`
}