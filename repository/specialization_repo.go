package repository

import (
	"service-antrik-chatbot/models"

	"gorm.io/gorm"
)

type SpecializationRepository interface {
	Create(spec *models.Specialization) error
	FindAll() ([]models.Specialization, error)
	FindByID(id uint) (*models.Specialization, error)
	Update(spec *models.Specialization) error
	Delete(id uint) error
}

type specializationRepository struct {
	db *gorm.DB
}

func NewSpecializationRepository(db *gorm.DB) SpecializationRepository {
	return &specializationRepository{db}
}

func (r *specializationRepository) Create(spec *models.Specialization) error {
	return r.db.Create(spec).Error
}

func (r *specializationRepository) FindAll() ([]models.Specialization, error) {
	var specs []models.Specialization
	err := r.db.Find(&specs).Error
	return specs, err
}

func (r *specializationRepository) FindByID(id uint) (*models.Specialization, error) {
	var spec models.Specialization
	err := r.db.First(&spec, id).Error
	if err != nil {
		return nil, err
	}
	return &spec, nil
}

func (r *specializationRepository) Update(spec *models.Specialization) error {
	return r.db.Save(spec).Error
}

func (r *specializationRepository) Delete(id uint) error {
	return r.db.Delete(&models.Specialization{}, id).Error
}
