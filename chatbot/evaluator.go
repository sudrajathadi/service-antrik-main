package chatbot

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"service-antrik-chatbot/models"
	"service-antrik-chatbot/repository"

	"gorm.io/gorm"
)

const (
	flowBooking          = "BOOKING"
	awaitingDoctor       = "DOCTOR"
	awaitingDate         = "DATE"
	awaitingTime         = "TIME"
	awaitingConfirmation = "CONFIRMATION"
)

type Evaluator struct {
	hospitals       repository.HospitalRepository
	specializations repository.SpecializationRepository
	doctors         repository.DoctorRepository
	schedules       repository.DoctorScheduleRepository
	users           repository.UserRepository
	appointments    repository.AppointmentRepository
	stateStore      StateStore
}

func NewEvaluator(
	hospitalRepo repository.HospitalRepository,
	specializationRepo repository.SpecializationRepository,
	doctorRepo repository.DoctorRepository,
	scheduleRepo repository.DoctorScheduleRepository,
	userRepo repository.UserRepository,
	appointmentRepo repository.AppointmentRepository,
	stateStore StateStore,
) *Evaluator {
	return &Evaluator{
		hospitals:       hospitalRepo,
		specializations: specializationRepo,
		doctors:         doctorRepo,
		schedules:       scheduleRepo,
		users:           userRepo,
		appointments:    appointmentRepo,
		stateStore:      stateStore,
	}
}

func (e *Evaluator) Evaluate(req ChatRequest, parsed ParseResult, intent Intent, confidence float64) (ChatResponse, error) {
	state := e.stateStore.Get(req.ChatID)
	state.ChatID = req.ChatID
	if req.UserID != 0 {
		state.UserID = uint(req.UserID)
	}
	e.enrichStateFromParsed(&state, parsed)

	response := ChatResponse{
		ChatID:     req.ChatID,
		Intent:     intent,
		Tokens:     parsed.Tokens,
		Parsed:     parsed,
		Confidence: confidence,
	}

	switch intent {
	case IntentGreeting:
		response.Reply = "Halo, saya bisa bantu cari dokter, cek jadwal, list rumah sakit, lokasi rumah sakit, spesialisasi, atau booking dokter."
	case IntentEmergency:
		response.Reply = "Keluhan ini bisa mengarah ke kondisi darurat. Segera ke IGD atau hubungi layanan darurat terdekat. Saya tidak akan membuat diagnosis, tetapi untuk gejala seperti ini sebaiknya jangan menunggu jadwal rawat jalan."
	case IntentCancelFlow:
		e.stateStore.Clear(req.ChatID)
		response.Reply = "Baik, alur saat ini saya batalkan. Kamu bisa mulai lagi dengan tanya dokter, jadwal, rumah sakit, spesialisasi, atau booking."
	case IntentConfirmBooking:
		return e.confirmBooking(req, state, response)
	case IntentListHospitals:
		return e.listHospitals(state, response)
	case IntentAskHospitalLocation:
		return e.hospitalLocation(state, response)
	case IntentListSpecializations:
		return e.listSpecializations(state, response)
	case IntentAskDoctor, IntentFindDoctorBySpecialization:
		return e.findDoctors(state, response)
	case IntentAskDoctorSchedule:
		return e.showSchedule(state, response)
	case IntentRecommendSpecialization:
		return e.recommendSpecialization(state, parsed, response)
	case IntentBookAppointment:
		return e.handleBooking(req, state, response)
	default:
		response.Reply = "Saya belum memahami pesan itu. Kamu bisa bertanya seperti: list rumah sakit, lokasi rumah sakit, jadwal dokter, dokter anak, atau booking dokter."
	}

	response.State = &state
	e.stateStore.Save(state)
	return response, nil
}

func (e *Evaluator) enrichStateFromParsed(state *ChatState, parsed ParseResult) {
	if parsed.Entities.Specialization != "" {
		state.SelectedSpecialty = parsed.Entities.Specialization
	}
	if parsed.Entities.Date != "" {
		state.SelectedDate = parsed.Entities.Date
	}
	if parsed.Entities.Time != "" {
		state.SelectedTime = parsed.Entities.Time
	}
	if len(parsed.Entities.Symptoms) > 0 {
		state.SymptomsNote = strings.Join(parsed.Entities.Symptoms, ", ")
	}
}

func (e *Evaluator) listHospitals(state ChatState, response ChatResponse) (ChatResponse, error) {
	hospitals, err := e.hospitals.FindAll()
	if err != nil {
		return response, err
	}
	location := strings.TrimSpace(response.Parsed.Entities.Location)
	if location != "" {
		hospitals = filterHospitalsByLocation(hospitals, location)
	}

	response.Data = hospitals
	if len(hospitals) == 0 {
		if location != "" {
			response.Reply = "Belum ada data rumah sakit yang tersedia untuk lokasi " + location + "."
		} else {
			response.Reply = "Belum ada data rumah sakit yang tersedia."
		}
	} else if location != "" {
		response.Reply = "Berikut daftar rumah sakit di " + location + ":\n" + joinHospitalNames(hospitals)
	} else {
		response.Reply = "Berikut daftar rumah sakit yang tersedia:\n" + joinHospitalNames(hospitals)
	}
	response.State = &state
	e.stateStore.Save(state)
	return response, nil
}

func (e *Evaluator) hospitalLocation(state ChatState, response ChatResponse) (ChatResponse, error) {
	hospitals, err := e.hospitals.FindAll()
	if err != nil {
		return response, err
	}

	hospital := matchHospital(hospitals, response.Parsed.Entities.HospitalName, response.Parsed.Entities.Location, response.Parsed.OriginalMessage)
	if hospital == nil {
		response.Reply = "Rumah sakit yang mana? Sebutkan nama rumah sakitnya agar saya bisa tampilkan alamatnya."
		response.NeedInput = []string{"hospital_name"}
		response.Data = hospitals
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	response.Data = hospital
	response.Reply = fmt.Sprintf("%s berlokasi di %s, %s. Nomor telepon: %s.", hospital.Name, hospital.Address, hospital.City, emptyDash(hospital.PhoneNumber))
	response.State = &state
	e.stateStore.Save(state)
	return response, nil
}

func (e *Evaluator) listSpecializations(state ChatState, response ChatResponse) (ChatResponse, error) {
	specs, err := e.specializations.FindAll()
	if err != nil {
		return response, err
	}

	response.Data = specs
	if len(specs) == 0 {
		response.Reply = "Belum ada data spesialisasi yang tersedia."
	} else {
		response.Reply = "Spesialisasi yang tersedia:\n" + joinSpecializationNames(specs)
	}
	response.State = &state
	e.stateStore.Save(state)
	return response, nil
}

func (e *Evaluator) findDoctors(state ChatState, response ChatResponse) (ChatResponse, error) {
	filter := repository.DoctorFilter{
		Specialization: state.SelectedSpecialty,
		Location:       response.Parsed.Entities.Location,
	}

	doctors, err := e.doctors.FindAllFiltered(filter)
	if err != nil {
		return response, err
	}

	if len(doctors) == 0 && state.SelectedSpecialty != "" {
		response.Reply = "Saya belum menemukan dokter untuk spesialisasi " + state.SelectedSpecialty + ". Coba pilih spesialisasi lain atau lihat daftar spesialisasi."
		response.NeedInput = []string{"specialization"}
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}
	if len(doctors) == 0 {
		response.Reply = "Dokter yang mana atau spesialisasi apa yang ingin dicari?"
		response.NeedInput = []string{"doctor_name", "specialization"}
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	summaries := summarizeDoctors(doctors)
	response.Data = summaries
	response.Reply = "Saya menemukan dokter berikut:\n" + joinDoctorNames(summaries) + "\n\nKamu bisa lanjut tanya jadwal atau booking dengan menyebut nama dokter."
	response.State = &state
	e.stateStore.Save(state)
	return response, nil
}

func (e *Evaluator) showSchedule(state ChatState, response ChatResponse) (ChatResponse, error) {
	doctor, err := e.resolveDoctor(state, response.Parsed)
	if err != nil {
		return response, err
	}
	if doctor == nil {
		response.Reply = "Jadwal dokter siapa yang ingin dilihat? Sebutkan nama dokter atau spesialisasinya."
		response.NeedInput = []string{"doctor_name"}
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	state.SelectedDoctorID = doctor.ID
	state.SelectedDoctorName = doctor.Name
	state.SelectedHospitalID = doctor.HospitalID
	state.SelectedHospitalName = doctor.Hospital.Name
	state.SelectedSpecialty = doctor.Specialization.Name

	schedules, err := e.schedules.FindAllByDoctorID(doctor.ID)
	if err != nil {
		return response, err
	}

	summaries := make([]ScheduleSummary, 0, len(schedules))
	for _, schedule := range schedules {
		if state.SelectedDate != "" {
			booked, err := e.schedules.GetBookedAppointments(schedule.DoctorID, state.SelectedDate)
			if err != nil {
				return response, err
			}
			schedule.TimeSlots = markBookedSlots(booked, schedule.TimeSlots)
		}
		summaries = append(summaries, ScheduleSummary{
			DoctorID:   doctor.ID,
			DoctorName: doctor.Name,
			DayOfWeek:  schedule.DayOfWeek,
			StartTime:  trimTime(schedule.StartTime),
			EndTime:    trimTime(schedule.EndTime),
			TimeSlots:  schedule.TimeSlots,
		})
	}

	response.Data = summaries
	if len(summaries) == 0 {
		response.Reply = doctor.Name + " belum memiliki jadwal praktik."
	} else if state.SelectedDate != "" {
		response.Reply = fmt.Sprintf("Jadwal %s untuk tanggal %s:\n%s\n\nSlot dengan status booked sudah terisi.", doctor.Name, state.SelectedDate, joinSchedules(summaries))
	} else {
		response.Reply = fmt.Sprintf("Jadwal %s:\n%s", doctor.Name, joinSchedules(summaries))
	}

	response.State = &state
	e.stateStore.Save(state)
	return response, nil
}

func (e *Evaluator) recommendSpecialization(state ChatState, parsed ParseResult, response ChatResponse) (ChatResponse, error) {
	recommended := recommendSpecialtyFromSymptoms(parsed.Entities.Symptoms)
	state.SelectedSpecialty = recommended

	if recommended == "" {
		response.Reply = "Keluhannya perlu diperjelas sedikit. Area keluhan utamanya di bagian apa, misalnya gigi, anak, kulit, mata, THT, jantung, atau perut?"
		response.NeedInput = []string{"symptoms"}
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	doctors, err := e.doctors.FindAllFiltered(repository.DoctorFilter{Specialization: recommended})
	if err != nil {
		return response, err
	}

	response.Data = summarizeDoctors(doctors)
	if len(doctors) == 0 {
		response.Reply = fmt.Sprintf("Untuk keluhan tersebut, biasanya bisa mulai dari spesialisasi %s. Namun saya belum menemukan dokter pada data saat ini.", recommended)
	} else {
		response.Reply = fmt.Sprintf("Untuk keluhan tersebut, biasanya bisa mulai dari spesialisasi %s.\n\nSaya menemukan dokter berikut:\n%s", recommended, joinDoctorNames(summarizeDoctors(doctors)))
	}

	response.State = &state
	e.stateStore.Save(state)
	return response, nil
}

func (e *Evaluator) handleBooking(req ChatRequest, state ChatState, response ChatResponse) (ChatResponse, error) {
	state.CurrentFlow = flowBooking

	if state.UserID == 0 {
		user, err := e.users.FindByChatID(req.ChatID)
		if err == nil {
			state.UserID = user.ID
		} else if err != gorm.ErrRecordNotFound {
			return response, err
		}
	}

	doctor, err := e.resolveDoctor(state, response.Parsed)
	if err != nil {
		return response, err
	}
	if doctor != nil {
		state.SelectedDoctorID = doctor.ID
		state.SelectedDoctorName = doctor.Name
		state.SelectedHospitalID = doctor.HospitalID
		state.SelectedHospitalName = doctor.Hospital.Name
		state.SelectedSpecialty = doctor.Specialization.Name
	}

	missing := missingBookingFields(state)
	if len(missing) > 0 {
		state.Awaiting = missing[0]
		response.NeedInput = missing
		response.Reply = bookingQuestion(state, missing[0])
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	state.Awaiting = awaitingConfirmation
	response.Reply = fmt.Sprintf(
		"Konfirmasi booking: %s di %s pada %s jam %s. Balas ya untuk membuat appointment atau batal untuk membatalkan.",
		state.SelectedDoctorName,
		state.SelectedHospitalName,
		state.SelectedDate,
		state.SelectedTime,
	)
	response.State = &state
	e.stateStore.Save(state)
	return response, nil
}

func (e *Evaluator) confirmBooking(req ChatRequest, state ChatState, response ChatResponse) (ChatResponse, error) {
	if state.CurrentFlow != flowBooking || state.Awaiting != awaitingConfirmation {
		response.Reply = "Belum ada booking yang menunggu konfirmasi. Mulai dengan: booking dokter anak besok jam 10:00."
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	missing := missingBookingFields(state)
	if len(missing) > 0 {
		response.NeedInput = missing
		response.Reply = bookingQuestion(state, missing[0])
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	appointmentDate, err := time.Parse("2006-01-02", state.SelectedDate)
	if err != nil {
		response.Reply = "Format tanggal belum valid. Gunakan format YYYY-MM-DD, contoh 2026-07-20."
		response.NeedInput = []string{awaitingDate}
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	appointment := models.Appointment{
		UserID:          state.UserID,
		DoctorID:        state.SelectedDoctorID,
		HospitalID:      state.SelectedHospitalID,
		AppointmentDate: appointmentDate,
		AppointmentTime: state.SelectedTime,
		SymptomsNote:    state.SymptomsNote,
		Status:          models.StatusPending,
	}

	if err := e.appointments.Create(&appointment); err != nil {
		response.Reply = "Booking belum bisa dibuat: " + err.Error()
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	e.stateStore.Clear(req.ChatID)
	response.Data = appointment
	response.State = nil
	response.Reply = fmt.Sprintf("Booking berhasil dibuat dengan status pending. Nomor appointment: %d.", appointment.ID)
	return response, nil
}

func (e *Evaluator) resolveDoctor(state ChatState, parsed ParseResult) (*models.Doctor, error) {
	if parsed.Entities.DoctorName != "" {
		doctors, err := e.doctors.FindAll()
		if err != nil {
			return nil, err
		}
		if doctor := matchDoctor(doctors, parsed.Entities.DoctorName, parsed.OriginalMessage); doctor != nil {
			return doctor, nil
		}
	}
	if state.SelectedDoctorID != 0 {
		return e.doctors.FindByID(state.SelectedDoctorID)
	}
	if state.SelectedSpecialty != "" {
		doctors, err := e.doctors.FindAllFiltered(repository.DoctorFilter{Specialization: state.SelectedSpecialty})
		if err != nil {
			return nil, err
		}
		if len(doctors) == 1 {
			return &doctors[0], nil
		}
	}
	return nil, nil
}

func missingBookingFields(state ChatState) []string {
	var missing []string
	if state.UserID == 0 {
		missing = append(missing, "user_id")
	}
	if state.SelectedDoctorID == 0 {
		missing = append(missing, awaitingDoctor)
	}
	if state.SelectedDate == "" {
		missing = append(missing, awaitingDate)
	}
	if state.SelectedTime == "" {
		missing = append(missing, awaitingTime)
	}
	return missing
}

func bookingQuestion(state ChatState, missing string) string {
	switch missing {
	case "user_id":
		return "Saya perlu user_id pasien untuk membuat appointment. Kirim ulang request chat dengan field user_id atau pastikan chat_id sudah terdaftar sebagai user."
	case awaitingDoctor:
		if state.SelectedSpecialty != "" {
			return "Dokter untuk spesialisasi " + state.SelectedSpecialty + " yang mana? Sebutkan nama dokter yang dipilih."
		}
		return "Dokter atau spesialisasi apa yang ingin dibooking?"
	case awaitingDate:
		return "Untuk tanggal berapa? Kamu bisa tulis hari ini, besok, lusa, atau format YYYY-MM-DD."
	case awaitingTime:
		return "Jam berapa? Gunakan format HH:MM, contoh 10:00."
	default:
		return "Mohon lengkapi data booking."
	}
}

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

func recommendSpecialtyFromSymptoms(symptoms []string) string {
	joined := strings.Join(symptoms, " ")
	switch {
	case strings.Contains(joined, "gigi"):
		return "gigi"
	case strings.Contains(joined, "anak"):
		return "anak"
	case strings.Contains(joined, "ruam") || strings.Contains(joined, "gatal"):
		return "kulit"
	case strings.Contains(joined, "mata"):
		return "mata"
	case strings.Contains(joined, "telinga") || strings.Contains(joined, "THT"):
		return "tht"
	case strings.Contains(joined, "nyeri dada"):
		return "jantung"
	case strings.Contains(joined, "sesak") || strings.Contains(joined, "batuk"):
		return "paru"
	case strings.Contains(joined, "hamil") || strings.Contains(joined, "menstruasi"):
		return "kandungan"
	case strings.Contains(joined, "kebas") || strings.Contains(joined, "kesemutan") || strings.Contains(joined, "sakit kepala"):
		return "saraf"
	case strings.Contains(joined, "perut") || strings.Contains(joined, "mual") || strings.Contains(joined, "muntah") || strings.Contains(joined, "diare"):
		return "penyakit dalam"
	default:
		return "umum"
	}
}

func matchDoctor(doctors []models.Doctor, candidate string, message string) *models.Doctor {
	candidate = normalizeMessage(candidate)
	message = normalizeMessage(message)
	for index := range doctors {
		name := normalizeMessage(doctors[index].Name)
		if candidate != "" && strings.Contains(name, candidate) {
			return &doctors[index]
		}
		if strings.Contains(message, name) {
			return &doctors[index]
		}
	}
	return nil
}

func matchHospital(hospitals []models.Hospital, candidates ...string) *models.Hospital {
	for _, candidate := range candidates {
		candidate = normalizeMessage(candidate)
		if candidate == "" {
			continue
		}
		for index := range hospitals {
			name := normalizeMessage(hospitals[index].Name)
			city := normalizeMessage(hospitals[index].City)
			if strings.Contains(name, candidate) || strings.Contains(candidate, name) || strings.Contains(city, candidate) {
				return &hospitals[index]
			}
		}
	}
	return nil
}

func filterHospitalsByLocation(hospitals []models.Hospital, location string) []models.Hospital {
	location = normalizeMessage(location)
	filtered := make([]models.Hospital, 0, len(hospitals))
	for _, hospital := range hospitals {
		name := normalizeMessage(hospital.Name)
		city := normalizeMessage(hospital.City)
		address := normalizeMessage(hospital.Address)
		if strings.Contains(name, location) || strings.Contains(city, location) || strings.Contains(address, location) {
			filtered = append(filtered, hospital)
		}
	}
	return filtered
}

func joinHospitalNames(hospitals []models.Hospital) string {
	names := make([]string, 0, len(hospitals))
	for _, hospital := range hospitals {
		names = append(names, "- "+hospital.Name+" ("+hospital.City+")")
	}
	sort.Strings(names)
	return strings.Join(names, "\n")
}

func joinSpecializationNames(specs []models.Specialization) string {
	names := make([]string, 0, len(specs))
	for _, spec := range specs {
		names = append(names, "- "+spec.Name)
	}
	sort.Strings(names)
	return strings.Join(names, "\n")
}

func joinDoctorNames(doctors []DoctorSummary) string {
	names := make([]string, 0, len(doctors))
	for _, doctor := range doctors {
		names = append(names, fmt.Sprintf("- %s (%s, %s)", doctor.Name, doctor.Specialization, doctor.Hospital))
	}
	sort.Strings(names)
	return strings.Join(names, "\n")
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

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
