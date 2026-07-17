package repository

import (
	"service-antrik-chatbot/models"
	"strings"

	"gorm.io/gorm"
)

type HospitalRepository interface {
	Create(hospital *models.Hospital) error
	FindAll() ([]models.Hospital, error)
	FindAllByCity(city string) ([]models.Hospital, error)
	FindByID(id uint) (*models.Hospital, error)
	Update(hospital *models.Hospital) error
	Delete(id uint) error
}

type hospitalRepository struct {
	db *gorm.DB
}

func NewHospitalRepository(db *gorm.DB) HospitalRepository {
	return &hospitalRepository{db}
}

func (r *hospitalRepository) Create(hospital *models.Hospital) error {
	return r.db.Create(hospital).Error
}

func (r *hospitalRepository) FindAll() ([]models.Hospital, error) {
	var hospitals []models.Hospital
	err := r.db.Find(&hospitals).Error
	return hospitals, err
}

func (r *hospitalRepository) FindAllByCity(city string) ([]models.Hospital, error) {
	city = strings.TrimSpace(city)
	if city == "" {
		return r.FindAll()
	}

	var hospitals []models.Hospital
	err := r.db.Where("city ILIKE ?", "%"+city+"%").Find(&hospitals).Error
	return hospitals, err
}

func (r *hospitalRepository) FindByID(id uint) (*models.Hospital, error) {
	var hospital models.Hospital
	err := r.db.First(&hospital, id).Error
	if err != nil {
		return nil, err
	}
	return &hospital, nil
}

func (r *hospitalRepository) Update(hospital *models.Hospital) error {
	return r.db.Save(hospital).Error
}

func (r *hospitalRepository) Delete(id uint) error {
	return r.db.Delete(&models.Hospital{}, id).Error
}
