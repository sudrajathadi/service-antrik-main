package evaluator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"service-antrik-chatbot/models"
)

var selectionNumberPattern = regexp.MustCompile(`\b\d+\b`)

func parseSelectionNumber(message string) (int, bool) {
	match := selectionNumberPattern.FindString(message)
	if match == "" {
		return 0, false
	}
	number, err := strconv.Atoi(match)
	if err != nil {
		return 0, false
	}
	return number, true
}

func nextDateForDay(dayName string, from time.Time) (string, bool) {
	target, ok := parseWeekday(dayName)
	if !ok {
		return "", false
	}

	base := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	daysAhead := (int(target) - int(base.Weekday()) + 7) % 7
	if daysAhead == 0 {
		daysAhead = 7
	}
	return base.AddDate(0, 0, daysAhead).Format("2006-01-02"), true
}

func parseWeekday(dayName string) (time.Weekday, bool) {
	switch strings.ToLower(strings.TrimSpace(dayName)) {
	case "sunday", "minggu":
		return time.Sunday, true
	case "monday", "senin":
		return time.Monday, true
	case "tuesday", "selasa":
		return time.Tuesday, true
	case "wednesday", "rabu":
		return time.Wednesday, true
	case "thursday", "kamis":
		return time.Thursday, true
	case "friday", "jumat", "jum'at":
		return time.Friday, true
	case "saturday", "sabtu":
		return time.Saturday, true
	default:
		return time.Sunday, false
	}
}

func buildTimeSlotOptions(schedule ScheduleOption, bookedAppointments []models.Appointment) []TimeSlotOption {
	start, errStart := parseClock(schedule.StartTime)
	end, errEnd := parseClock(schedule.EndTime)
	if errStart != nil || errEnd != nil || schedule.SlotInterval <= 0 {
		return nil
	}

	booked := make(map[string]bool)
	for _, appointment := range bookedAppointments {
		booked[trimTime(appointment.AppointmentTime)] = true
	}

	var options []TimeSlotOption
	for current := start; current.Before(end); current = current.Add(time.Duration(schedule.SlotInterval) * time.Minute) {
		value := current.Format("15:04")
		if booked[value] {
			continue
		}
		options = append(options, TimeSlotOption{
			Number: len(options) + 1,
			Date:   schedule.Date,
			Time:   value,
			Booked: false,
		})
	}
	return options
}

func scheduleMatchesDate(dayName string, date string) (bool, error) {
	scheduleDay, ok := parseWeekday(dayName)
	if !ok {
		return false, nil
	}
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false, err
	}
	return parsedDate.Weekday() == scheduleDay, nil
}

func parseClock(value string) (time.Time, error) {
	for _, layout := range []string{"15:04", "15:04:05"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time %q", value)
}

func markBookedSlots(bookedAppointments []models.Appointment, allSlots []models.TimeSlot) []models.TimeSlot {
	bookedTimes := make(map[string]bool)
	for _, appointment := range bookedAppointments {
		bookedTimes[trimTime(appointment.AppointmentTime)] = true
	}
	for index := range allSlots {
		if bookedTimes[allSlots[index].Time] {
			allSlots[index].Booked = true
		}
	}
	return allSlots
}

func trimTime(value string) string {
	for _, layout := range []string{"15:04:05", "15:04"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.Format("15:04")
		}
	}
	return value
}
