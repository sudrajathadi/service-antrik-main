package repository

import (
	"service-antrik-chatbot/models"
	"strings"

	"gorm.io/gorm"
)

type DoctorFilter struct {
	Specialization string
	City           string
	Location       string
}

type DoctorRepository interface {
	Create(doctor *models.Doctor) error
	FindAll() ([]models.Doctor, error)
	FindAllFiltered(filter DoctorFilter) ([]models.Doctor, error)
	FindByID(id uint) (*models.Doctor, error)
	Update(doctor *models.Doctor) error
	Delete(id uint) error
}

type doctorRepository struct {
	db *gorm.DB
}

func NewDoctorRepository(db *gorm.DB) DoctorRepository {
	return &doctorRepository{db}
}

func (r *doctorRepository) Create(doctor *models.Doctor) error {
	return r.db.Create(doctor).Error
}

func (r *doctorRepository) FindAll() ([]models.Doctor, error) {
	return r.FindAllFiltered(DoctorFilter{})
}

func (r *doctorRepository) FindAllFiltered(filter DoctorFilter) ([]models.Doctor, error) {
	var doctors []models.Doctor
	query := r.db.Model(&models.Doctor{}).
		Preload("Specialization").
		Preload("Hospital").
		Joins("JOIN specializations ON specializations.id = doctors.specialization_id").
		Joins("JOIN hospitals ON hospitals.id = doctors.hospital_id")

	if specialization := strings.TrimSpace(filter.Specialization); specialization != "" {
		query = query.Where("specializations.name ILIKE ?", "%"+specialization+"%")
	}

	if city := strings.TrimSpace(filter.City); city != "" {
		query = query.Where("hospitals.city ILIKE ?", "%"+city+"%")
	}

	if location := strings.TrimSpace(filter.Location); location != "" {
		value := "%" + location + "%"
		query = query.Where(
			"hospitals.city ILIKE ? OR hospitals.name ILIKE ? OR hospitals.address ILIKE ?",
			value,
			value,
			value,
		)
	}

	err := query.Find(&doctors).Error
	return doctors, err
}

func (r *doctorRepository) FindByID(id uint) (*models.Doctor, error) {
	var doctor models.Doctor
	err := r.db.Preload("Specialization").Preload("Hospital").First(&doctor, id).Error
	if err != nil {
		return nil, err
	}
	return &doctor, nil
}

func (r *doctorRepository) Update(doctor *models.Doctor) error {
	return r.db.Save(doctor).Error
}

func (r *doctorRepository) Delete(id uint) error {
	return r.db.Delete(&models.Doctor{}, id).Error
}
