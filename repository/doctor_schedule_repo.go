package repository

import (
	"doctor-booking/models"

	"gorm.io/gorm"
)

type DoctorScheduleRepository interface {
	Create(schedule *models.DoctorSchedule) error
	FindAll() ([]models.DoctorSchedule, error)
	FindByID(id uint) (*models.DoctorSchedule, error)
	Update(schedule *models.DoctorSchedule) error
	Delete(id uint) error
}

type doctorScheduleRepository struct {
	db *gorm.DB
}

func NewDoctorScheduleRepository(db *gorm.DB) DoctorScheduleRepository {
	return &doctorScheduleRepository{db}
}

func (r *doctorScheduleRepository) Create(schedule *models.DoctorSchedule) error {
	return r.db.Create(schedule).Error
}

func (r *doctorScheduleRepository) FindAll() ([]models.DoctorSchedule, error) {
	var schedules []models.DoctorSchedule
	err := r.db.Preload("Doctor").Find(&schedules).Error
	return schedules, err
}

func (r *doctorScheduleRepository) FindByID(id uint) (*models.DoctorSchedule, error) {
	var schedule models.DoctorSchedule
	err := r.db.Preload("Doctor").First(&schedule, id).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *doctorScheduleRepository) Update(schedule *models.DoctorSchedule) error {
	return r.db.Save(schedule).Error
}

func (r *doctorScheduleRepository) Delete(id uint) error {
	return r.db.Delete(&models.DoctorSchedule{}, id).Error
}
