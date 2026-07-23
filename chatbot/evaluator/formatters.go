package evaluator

import (
	"fmt"
	"sort"
	"strings"

	"service-antrik-chatbot/models"
)

func summarizeDoctors(doctors []models.Doctor) []DoctorSummary {
	summaries := make([]DoctorSummary, 0, len(doctors))
	for _, doctor := range doctors {
		if !doctor.IsActive {
			continue
		}
		summaries = append(summaries, DoctorSummary{
			ID:             doctor.ID,
			Name:           doctor.Name,
			Specialization: doctor.Specialization.Name,
			HospitalID:     doctor.HospitalID,
			Hospital:       doctor.Hospital.Name,
			City:           doctor.Hospital.City,
			Experience:     doctor.ExperienceYears,
		})
	}
	return summaries
}

func summarizeHospitals(hospitals []models.Hospital) []HospitalSummary {
	summaries := make([]HospitalSummary, 0, len(hospitals))
	for _, hospital := range hospitals {
		summaries = append(summaries, HospitalSummary{
			ID:          hospital.ID,
			Name:        hospital.Name,
			Address:     hospital.Address,
			City:        hospital.City,
			PhoneNumber: hospital.PhoneNumber,
		})
	}
	return summaries
}

func joinHospitalNames(hospitals []models.Hospital) string {
	names := make([]string, 0, len(hospitals))
	for _, hospital := range hospitals {
		names = append(names, fmt.Sprintf("- %s (%s) - %s", hospital.Name, hospital.City, hospital.Address))
	}
	sort.Strings(names)
	return strings.Join(names, "\n")
}

func joinNumberedHospitalNames(hospitals []HospitalSummary) string {
	lines := make([]string, 0, len(hospitals))
	for index, hospital := range hospitals {
		lines = append(lines, fmt.Sprintf("%d. %s (%s) - %s", index+1, hospital.Name, hospital.City, hospital.Address))
	}
	return strings.Join(lines, "\n")
}

func joinSpecializationNames(specs []models.Specialization) string {
	names := make([]string, 0, len(specs))
	for _, spec := range specs {
		names = append(names, "- "+spec.Name)
	}
	sort.Strings(names)
	return strings.Join(names, "\n")
}

func joinNumberedDoctorNames(doctors []DoctorSummary) string {
	lines := make([]string, 0, len(doctors))
	for index, doctor := range doctors {
		lines = append(lines, fmt.Sprintf("%d. %s (%s, %s)", index+1, doctor.Name, doctor.Specialization, doctor.Hospital))
	}
	return strings.Join(lines, "\n")
}

func joinSchedules(schedules []ScheduleSummary) string {
	parts := make([]string, 0, len(schedules))
	for _, schedule := range schedules {
		line := fmt.Sprintf("- %s %s-%s", schedule.DayOfWeek, trimTime(schedule.StartTime), trimTime(schedule.EndTime))
		if len(schedule.TimeSlots) > 0 {
			line += "\n  Slot:\n" + joinTimeSlots(schedule.TimeSlots)
		}
		parts = append(parts, line)
	}
	return strings.Join(parts, "\n")
}

func joinNumberedScheduleOptions(options []ScheduleOption) string {
	lines := make([]string, 0, len(options))
	for _, option := range options {
		lines = append(lines, fmt.Sprintf("%d. %s, %s (%s-%s)", option.Number, option.DayOfWeek, option.Date, option.StartTime, option.EndTime))
	}
	return strings.Join(lines, "\n")
}

func joinTimeSlots(slots []models.TimeSlot) string {
	lines := make([]string, 0, len(slots))
	for _, slot := range slots {
		status := "available"
		if slot.Booked {
			status = "booked"
		}
		lines = append(lines, fmt.Sprintf("  - %s (%s)", slot.Time, status))
	}
	return strings.Join(lines, "\n")
}

func joinNumberedTimeSlotOptions(options []TimeSlotOption) string {
	lines := make([]string, 0, len(options))
	for _, option := range options {
		lines = append(lines, fmt.Sprintf("%d. %s", option.Number, option.Time))
	}
	return strings.Join(lines, "\n")
}

func bookingSuccessMessage(appointment models.Appointment, state ChatState, patientName string, patientPhone string, patientEmail string) string {
	lines := []string{
		"Booking berhasil dibuat dengan status pending.",
		fmt.Sprintf("Nomor appointment: %d", appointment.ID),
		"Dokter: " + state.SelectedDoctorName,
		"Rumah sakit: " + state.SelectedHospitalName,
		"Tanggal: " + state.SelectedDate,
		"Jam: " + state.SelectedTime,
		"Keluhan: " + emptyDash(state.PatientComplaint),
		"",
		"Data pasien:",
		"Nama: " + emptyDash(patientName),
		"Telepon: " + emptyDash(patientPhone),
		"Email: " + emptyDash(patientEmail),
	}
	return strings.Join(lines, "\n")
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
