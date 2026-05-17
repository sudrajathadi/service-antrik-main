package models

type Doctor struct {
	Base
	SpecializationID uint           `json:"specialization_id" gorm:"not null;index"`
	HospitalID       uint           `json:"hospital_id"       gorm:"not null;index"`
	Name             string         `json:"name"              gorm:"type:varchar(255);not null"`
	Bio              string         `json:"bio"               gorm:"type:text"`
	ExperienceYears  int            `json:"experience_years"  gorm:"default:0"`
	IsActive         bool           `json:"is_active"         gorm:"default:true"`
	Specialization   Specialization `json:"specialization"    gorm:"foreignKey:SpecializationID"`
	Hospital         Hospital       `json:"hospital"          gorm:"foreignKey:HospitalID"`
}
