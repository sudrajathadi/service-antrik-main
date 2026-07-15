package repository

import (
	"errors"
	"service-antrik-chatbot/models"

	"gorm.io/gorm"
)

type DoctorScheduleRepository interface {
	Create(schedule *models.DoctorSchedule) error
	FindAll() ([]models.DoctorSchedule, error)
	FindAllByDoctorID(doctorID uint) ([]models.DoctorSchedule, error)
	FindByID(id uint) (*models.DoctorSchedule, error)
	Update(schedule *models.DoctorSchedule) error
	Delete(id uint) error
	GetBookedAppointments(doctorID uint, dateStr string) ([]models.Appointment, error)
}

type doctorScheduleRepository struct {
	db *gorm.DB
}

func NewDoctorScheduleRepository(db *gorm.DB) DoctorScheduleRepository {
	return &doctorScheduleRepository{db}
}

func (r *doctorScheduleRepository) GetBookedAppointments(doctorID uint, dateStr string) ([]models.Appointment, error) {
	var bookedAppointments []models.Appointment

	err := r.db.Where("doctor_id = ? AND DATE(appointment_date) = ? AND status != ?",
		doctorID, dateStr, models.StatusCancelled).
		Find(&bookedAppointments).Error

	return bookedAppointments, err
}

func (r *doctorScheduleRepository) Create(schedule *models.DoctorSchedule) error {
	// 1. Basic sanity check: Start time must be before End time
	if schedule.StartTime >= schedule.EndTime {
		return errors.New("start time must be before end time")
	}

	// 2. Check for overlapping shifts
	var count int64
	err := r.db.Model(&models.DoctorSchedule{}).
		Where("doctor_id = ? AND day_of_week = ?", schedule.DoctorID, schedule.DayOfWeek).
		Where("start_time < ? AND end_time > ?", schedule.EndTime, schedule.StartTime).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("schedule conflict: this shift overlaps with an existing shift on this day")
	}

	// 3. If no conflicts, safely insert into DB
	return r.db.Create(schedule).Error
}

func (r *doctorScheduleRepository) FindAll() ([]models.DoctorSchedule, error) {
	var schedules []models.DoctorSchedule
	err := r.db.Preload("Doctor").Find(&schedules).Error
	return schedules, err
}

func (r *doctorScheduleRepository) FindAllByDoctorID(doctorID uint) ([]models.DoctorSchedule, error) {
	var schedules []models.DoctorSchedule
	err := r.db.Preload("Doctor").Where("doctor_id = ?", doctorID).Find(&schedules).Error
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
