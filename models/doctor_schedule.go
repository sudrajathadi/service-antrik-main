package models

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Add this new struct to represent the JSON object
type TimeSlot struct {
	Time   string `json:"time"`
	Booked bool   `json:"booked"`
}
type DoctorSchedule struct {
	Base
	DoctorID     uint       `json:"doctor_id" gorm:"not null;index"`
	DayOfWeek    string     `json:"day_of_week" gorm:"type:day_name;not null"`

	StartTime    string     `json:"start_time" gorm:"type:time without time zone;not null"`
	EndTime      string     `json:"end_time" gorm:"type:time without time zone;not null"`

	SlotInterval int        `json:"slot_interval" gorm:"default:30"`
	TimeSlots    []TimeSlot `json:"time_slots" gorm:"-"`
	Doctor       Doctor     `json:"doctor" gorm:"foreignKey:DoctorID"`
}

// GenerateSlots now returns a slice of TimeSlot structs
func (ds *DoctorSchedule) GenerateSlots() {
	if ds.SlotInterval <= 0 || ds.StartTime == "" || ds.EndTime == "" {
		return
	}

	layoutFull := "15:04:05"
	layoutShort := "15:04"

	start, errStart := time.Parse(layoutFull, ds.StartTime)
	if errStart != nil {
		start, errStart = time.Parse(layoutShort, ds.StartTime)
	}

	end, errEnd := time.Parse(layoutFull, ds.EndTime)
	if errEnd != nil {
		end, errEnd = time.Parse(layoutShort, ds.EndTime)
	}

	if errStart != nil || errEnd != nil {
		return
	}

	// Initialize the slice
	slots := make([]TimeSlot, 0)
	interval := time.Duration(ds.SlotInterval) * time.Minute

	// Create objects and default Booked to false
	for current := start; current.Before(end); current = current.Add(interval) {
		slots = append(slots, TimeSlot{
			Time:   current.Format("15:04"),
			Booked: false, 
		})
	}

	ds.TimeSlots = slots
}

// (Keep AfterFind, AfterSave, and BeforeSave exactly as they were)
func (ds *DoctorSchedule) AfterFind(tx *gorm.DB) error {
	ds.GenerateSlots()
	return nil
}

func (ds *DoctorSchedule) AfterSave(tx *gorm.DB) error {
	if ds.ID != 0 {
		ds.GenerateSlots()
	}
	return nil
}

// BeforeSave contains all our validation rules before touching the database
func (ds *DoctorSchedule) BeforeSave(tx *gorm.DB) error {
	layoutFull := "15:04:05"
	layoutShort := "15:04"

	parsedStart, err := time.Parse(layoutFull, ds.StartTime)
	if err != nil {
		parsedStart, err = time.Parse(layoutShort, ds.StartTime)
		if err != nil {
			return errors.New("invalid start_time format. Use HH:MM or HH:MM:SS")
		}
	}

	parsedEnd, err := time.Parse(layoutFull, ds.EndTime)
	if err != nil {
		parsedEnd, err = time.Parse(layoutShort, ds.EndTime)
		if err != nil {
			return errors.New("invalid end_time format. Use HH:MM or HH:MM:SS")
		}
	}

	ds.StartTime = parsedStart.Format(layoutFull)
	ds.EndTime = parsedEnd.Format(layoutFull)

	if !parsedStart.Before(parsedEnd) {
		return errors.New("start time must be earlier than end time")
	}

	if ds.SlotInterval <= 0 {
		return errors.New("slot interval must be greater than zero")
	}

	shiftDuration := parsedEnd.Sub(parsedStart)
	durationMinutes := int(shiftDuration.Minutes())

	if durationMinutes%ds.SlotInterval != 0 {
		return fmt.Errorf("shift duration (%d mins) does not match slot interval (%d mins). Times must align perfectly", durationMinutes, ds.SlotInterval)
	}

	var count int64
	query := tx.Model(&DoctorSchedule{}).
		Where("doctor_id = ?", ds.DoctorID).
		Where("day_of_week = ?", ds.DayOfWeek).
		Where("start_time < ? AND end_time > ?", ds.EndTime, ds.StartTime)

	if ds.ID != 0 {
		query = query.Where("id != ?", ds.ID)
	}

	if err := query.Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return errors.New("schedule overlaps with an existing shift on this day")
	}

	return nil
}