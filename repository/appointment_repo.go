package repository

import (
	"errors"
	"fmt"
	"service-antrik-chatbot/models"
	"time"

	"gorm.io/gorm"
)

type AppointmentRepository interface {
	Create(appointment *models.Appointment) error
	FindAll() ([]models.Appointment, error)
	FindByID(id uint) (*models.Appointment, error)
	Update(appointment *models.Appointment) error
	Delete(id uint) error
}

type appointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepository{db}
}

func (r *appointmentRepository) Create(appointment *models.Appointment) error {
	if err := r.validateCreate(appointment); err != nil {
		return err
	}
	return r.db.Create(appointment).Error
}

func (r *appointmentRepository) validateCreate(appointment *models.Appointment) error {
	if appointment.UserID == 0 {
		return errors.New("user_id is required and must reference an existing user")
	}
	if appointment.DoctorID == 0 {
		return errors.New("doctor_id is required and must reference an existing doctor")
	}
	if appointment.HospitalID == 0 {
		return errors.New("hospital_id is required and must reference an existing hospital")
	}
	if appointment.AppointmentDate.IsZero() {
		return errors.New("appointment_date is required")
	}
	if appointment.AppointmentTime == "" {
		return errors.New("appointment_time is required")
	}
	appointmentTime, err := parseAppointmentTime(appointment.AppointmentTime)
	if err != nil {
		return errors.New("appointment_time must use HH:MM format")
	}
	appointment.AppointmentTime = appointmentTime.Format("15:04")

	if err := r.ensureExists(&models.User{}, appointment.UserID, "user_id"); err != nil {
		return err
	}
	if err := r.ensureExists(&models.Hospital{}, appointment.HospitalID, "hospital_id"); err != nil {
		return err
	}

	var doctor models.Doctor
	if err := r.db.First(&doctor, appointment.DoctorID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("doctor_id %d was not found", appointment.DoctorID)
		}
		return err
	}
	if doctor.HospitalID != appointment.HospitalID {
		return fmt.Errorf("doctor_id %d does not practice at hospital_id %d", appointment.DoctorID, appointment.HospitalID)
	}

	dayName := appointment.AppointmentDate.Weekday().String()
	var schedule models.DoctorSchedule
	if err := r.db.
		Where("doctor_id = ? AND LOWER(day_of_week::text) = LOWER(?)", appointment.DoctorID, dayName).
		First(&schedule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("doctor_id %d has no schedule on %s", appointment.DoctorID, dayName)
		}
		return err
	}

	startTime, errStart := parseAppointmentTime(schedule.StartTime)
	endTime, errEnd := parseAppointmentTime(schedule.EndTime)
	if errStart != nil || errEnd != nil {
		return errors.New("doctor schedule has invalid start_time or end_time")
	}
	if schedule.SlotInterval <= 0 {
		return errors.New("doctor schedule has invalid slot_interval")
	}
	if appointmentTime.Before(startTime) || !appointmentTime.Before(endTime) {
		return fmt.Errorf("appointment_time %s is outside schedule %s-%s", appointment.AppointmentTime, startTime.Format("15:04"), endTime.Format("15:04"))
	}
	if int(appointmentTime.Sub(startTime).Minutes())%schedule.SlotInterval != 0 {
		return fmt.Errorf("appointment_time %s does not align with slot interval %d minutes", appointment.AppointmentTime, schedule.SlotInterval)
	}

	var bookedCount int64
	if err := r.db.Model(&models.Appointment{}).
		Where("doctor_id = ? AND DATE(appointment_date) = DATE(?) AND appointment_time = ? AND status != ?",
			appointment.DoctorID,
			appointment.AppointmentDate,
			appointment.AppointmentTime,
			models.StatusCancelled,
		).
		Count(&bookedCount).Error; err != nil {
		return err
	}
	if bookedCount > 0 {
		return fmt.Errorf("slot %s is already booked for this doctor on %s", appointment.AppointmentTime, appointment.AppointmentDate.Format("2006-01-02"))
	}

	return nil
}

func (r *appointmentRepository) ensureExists(model interface{}, id uint, field string) error {
	var count int64
	if err := r.db.Model(model).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("%s %d was not found", field, id)
	}
	return nil
}

func parseAppointmentTime(value string) (time.Time, error) {
	for _, layout := range []string{"15:04", "15:04:05"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("invalid time")
}

func (r *appointmentRepository) FindAll() ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Preload("User").Preload("Doctor").Preload("Hospital").Find(&appointments).Error
	return appointments, err
}

func (r *appointmentRepository) FindByID(id uint) (*models.Appointment, error) {
	var appointment models.Appointment
	err := r.db.Preload("User").Preload("Doctor").Preload("Hospital").First(&appointment, id).Error
	if err != nil {
		return nil, err
	}
	return &appointment, nil
}

func (r *appointmentRepository) Update(appointment *models.Appointment) error {
	return r.db.Save(appointment).Error
}

func (r *appointmentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Appointment{}, id).Error
}
