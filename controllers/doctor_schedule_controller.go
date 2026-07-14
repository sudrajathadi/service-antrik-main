package controllers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"service-antrik-chatbot/models"
	"service-antrik-chatbot/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DoctorScheduleController struct {
	repo repository.DoctorScheduleRepository
}

func NewDoctorScheduleController(repo repository.DoctorScheduleRepository) *DoctorScheduleController {
	return &DoctorScheduleController{repo}
}

const ErrMsgInvalidID = "invalid id"

func (c *DoctorScheduleController) Create(ctx *gin.Context) {
	var schedule models.DoctorSchedule
	if err := ctx.ShouldBindJSON(&schedule); err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid schedule request body", err.Error())
		return
	}
	if err := c.repo.Create(&schedule); err != nil {
		respondError(ctx, http.StatusUnprocessableEntity, "SCHEDULE_CREATE_FAILED", "Schedule could not be created", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusCreated, "Schedule created successfully", schedule)
}

// Helper function: Now it just accepts the raw data instead of querying the DB itself
func markBookedSlots(bookedAppointments []models.Appointment, allSlots []models.TimeSlot) []models.TimeSlot {
	// Create map for O(1) lookups
	bookedTimes := make(map[string]bool)
	for _, appt := range bookedAppointments {
		parsedTime, err := time.Parse("15:04:05", appt.AppointmentTime)
		if err != nil {
			parsedTime, _ = time.Parse("15:04", appt.AppointmentTime)
		}
		bookedTimes[parsedTime.Format("15:04")] = true
	}

	// Loop through reference array and flip the bool if it exists in our map
	for i := range allSlots {
		if bookedTimes[allSlots[i].Time] {
			allSlots[i].Booked = true
		}
	}

	return allSlots
}

func generateTimeSlots(start, end string, interval int) []models.TimeSlot {
	layout := "15:04"

	startTime, errStart := parseScheduleTime(start)
	endTime, errEnd := parseScheduleTime(end)
	if errStart != nil || errEnd != nil || interval <= 0 {
		return nil
	}

	var slots []models.TimeSlot

	for startTime.Before(endTime) {
		slots = append(slots, models.TimeSlot{
			Time:   startTime.Format(layout),
			Booked: false,
		})

		startTime = startTime.Add(time.Minute * time.Duration(interval))
	}

	return slots
}

func parseScheduleTime(value string) (time.Time, error) {
	for _, layout := range []string{"15:04:05", "15:04"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, errors.New("invalid time format")
}

func (c *DoctorScheduleController) GetAll(ctx *gin.Context) {
	schedules, err := c.repo.FindAll()

	if err != nil {
		log.Println("ERROR FIND ALL:", err)

		respondError(ctx, http.StatusInternalServerError, "SCHEDULES_FETCH_FAILED", "Schedules could not be fetched", err.Error())
		return
	}

	log.Println("TOTAL SCHEDULES:", len(schedules))

	dateStr := ctx.Query("date")
	log.Println("DATE QUERY:", dateStr)

	var mappedSchedules []models.DoctorSchedule

	for index, schedule := range schedules {

		log.Println("===================================")
		log.Println("LOOP INDEX:", index)
		log.Println("DOCTOR ID:", schedule.DoctorID)
		log.Println("DAY:", schedule.DayOfWeek)
		log.Println("START:", schedule.StartTime)
		log.Println("END:", schedule.EndTime)
		log.Println("INTERVAL:", schedule.SlotInterval)

		generatedSlots := generateTimeSlots(
			schedule.StartTime,
			schedule.EndTime,
			schedule.SlotInterval,
		)

		log.Println("GENERATED SLOTS COUNT:", len(generatedSlots))
		log.Printf("GENERATED SLOTS: %+v\n", generatedSlots)

		schedule.TimeSlots = generatedSlots

		if dateStr != "" {

			bookedAppointments, err := c.repo.GetBookedAppointments(
				schedule.DoctorID,
				dateStr,
			)

			if err != nil {
				log.Println("ERROR GET BOOKED APPOINTMENTS:", err)
			}

			log.Println("BOOKED APPOINTMENTS COUNT:", len(bookedAppointments))
			log.Printf("BOOKED APPOINTMENTS: %+v\n", bookedAppointments)

			schedule.TimeSlots = markBookedSlots(
				bookedAppointments,
				schedule.TimeSlots,
			)

			log.Printf("TIME SLOTS AFTER BOOKING: %+v\n", schedule.TimeSlots)
		}

		log.Printf("FINAL SCHEDULE: %+v\n", schedule)

		mappedSchedules = append(mappedSchedules, schedule)
	}

	log.Println("===================================")
	log.Println("FINAL MAPPED SCHEDULES COUNT:", len(mappedSchedules))
	log.Printf("FINAL RESPONSE: %+v\n", mappedSchedules)

	respondSuccess(ctx, http.StatusOK, "Schedules fetched successfully", mappedSchedules)
}

func (c *DoctorScheduleController) GetByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_SCHEDULE_ID", "Invalid schedule id", err.Error())
		return
	}

	schedule, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(ctx, http.StatusNotFound, "SCHEDULE_NOT_FOUND", "Schedule not found", err.Error())
			return
		}
		respondError(ctx, http.StatusInternalServerError, "SCHEDULE_FETCH_FAILED", "Schedule could not be fetched", err.Error())
		return
	}

	dateStr := ctx.Query("date")

	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			respondError(ctx, http.StatusBadRequest, "INVALID_DATE", "Invalid date format. Use YYYY-MM-DD", err.Error())
			return
		}

		requestedDay := strings.ToLower(parsedDate.Weekday().String())
		if requestedDay != strings.ToLower(schedule.DayOfWeek) {
			respondError(ctx, http.StatusBadRequest, "DATE_DAY_MISMATCH", "The requested date does not match the schedule's day of the week", "requested day is "+requestedDay+", schedule day is "+schedule.DayOfWeek)
			return
		}

		// 1. Fetch appointments from your repository
		bookedAppointments, _ := c.repo.GetBookedAppointments(schedule.DoctorID, dateStr)

		// 2. Mark the slots as booked
		generatedSlots := generateTimeSlots(
			schedule.StartTime,
			schedule.EndTime,
			schedule.SlotInterval,
		)

		schedule.TimeSlots = markBookedSlots(
			bookedAppointments,
			generatedSlots,
		)
	}

	respondSuccess(ctx, http.StatusOK, "Schedule fetched successfully", schedule)
}

func (c *DoctorScheduleController) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_SCHEDULE_ID", "Invalid schedule id", err.Error())
		return
	}
	schedule, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(ctx, http.StatusNotFound, "SCHEDULE_NOT_FOUND", "Schedule not found", err.Error())
			return
		}
		respondError(ctx, http.StatusInternalServerError, "SCHEDULE_FETCH_FAILED", "Schedule could not be fetched", err.Error())
		return
	}
	if err := ctx.ShouldBindJSON(schedule); err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid schedule request body", err.Error())
		return
	}
	if err := c.repo.Update(schedule); err != nil {
		respondError(ctx, http.StatusInternalServerError, "SCHEDULE_UPDATE_FAILED", "Schedule could not be updated", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Schedule updated successfully", schedule)
}

func (c *DoctorScheduleController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_SCHEDULE_ID", "Invalid schedule id", err.Error())
		return
	}
	if err := c.repo.Delete(uint(id)); err != nil {
		respondError(ctx, http.StatusInternalServerError, "SCHEDULE_DELETE_FAILED", "Schedule could not be deleted", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Schedule deleted successfully", gin.H{"id": id})
}
