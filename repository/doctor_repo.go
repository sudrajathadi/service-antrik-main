package repository

import (
	"service-antrik-chatbot/models"

	"gorm.io/gorm"
)

type DoctorRepository interface {
	Create(doctor *models.Doctor) error
	FindAll() ([]models.Doctor, error)
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
	var doctors []models.Doctor
	err := r.db.Preload("Specialization").Preload("Hospital").Find(&doctors).Error
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
