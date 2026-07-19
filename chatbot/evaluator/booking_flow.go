package evaluator

import (
	"fmt"
	"time"

	"service-antrik-chatbot/models"
	"service-antrik-chatbot/repository"

	"gorm.io/gorm"
)

const (
	flowBooking               = "BOOKING"
	awaitingDoctor            = "DOCTOR"
	awaitingDoctorSelection   = "DOCTOR_SELECTION"
	awaitingScheduleSelection = "SCHEDULE_SELECTION"
	awaitingTimeSelection     = "TIME_SELECTION"
	awaitingPatientDetails    = "PATIENT_DETAILS"
)

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
		rememberDoctor(&state, *doctor)
		return e.showBookableScheduleOptions(state, response)
	}

	if state.SelectedSpecialty != "" {
		doctors, err := e.doctors.FindAllFiltered(repository.DoctorFilter{Specialization: state.SelectedSpecialty})
		if err != nil {
			return response, err
		}
		if len(doctors) > 0 {
			summaries := summarizeDoctors(doctors)
			state.PendingDoctors = summaries
			state.Awaiting = awaitingDoctorSelection
			response.Data = summaries
			response.NeedInput = []string{awaitingDoctorSelection}
			response.Reply = "Pilih dokter untuk spesialisasi " + state.SelectedSpecialty + ":\n" + joinNumberedDoctorNames(summaries) + "\n\nBalas dengan nomor dokter."
			return e.finish(response, state)
		}
	}

	state.Awaiting = awaitingDoctor
	response.NeedInput = []string{"doctor_name", "specialization"}
	response.Reply = "Dokter atau spesialisasi apa yang ingin dibooking?"
	return e.finish(response, state)
}

func (e *Evaluator) continueBookingFlow(req ChatRequest, state ChatState, response ChatResponse) (bool, ChatResponse, error) {
	if state.CurrentFlow != flowBooking {
		return false, response, nil
	}

	switch state.Awaiting {
	case awaitingDoctor:
		flowResponse, err := e.handleBooking(req, state, response)
		return true, flowResponse, err
	case awaitingDoctorSelection:
		number, ok := parseSelectionNumber(req.Message)
		if !ok {
			return false, response, nil
		}
		flowResponse, err := e.selectDoctorByNumber(state, response, number)
		return true, flowResponse, err
	case awaitingScheduleSelection:
		number, ok := parseSelectionNumber(req.Message)
		if !ok {
			return false, response, nil
		}
		flowResponse, err := e.selectScheduleByNumber(state, response, number)
		return true, flowResponse, err
	case awaitingTimeSelection:
		number, ok := parseSelectionNumber(req.Message)
		if !ok {
			return false, response, nil
		}
		flowResponse, err := e.selectTimeByNumber(state, response, number)
		return true, flowResponse, err
	case awaitingPatientDetails:
		flowResponse, err := e.createAppointmentFromPatientDetails(req, state, response)
		return true, flowResponse, err
	default:
		return false, response, nil
	}
}

func (e *Evaluator) selectDoctorByNumber(state ChatState, response ChatResponse, number int) (ChatResponse, error) {
	if number < 1 || number > len(state.PendingDoctors) {
		return e.replyWithState(response, state, fmt.Sprintf("Nomor dokter tidak tersedia. Pilih nomor 1 sampai %d.", len(state.PendingDoctors)))
	}

	selected := state.PendingDoctors[number-1]
	rememberDoctorSummary(&state, selected)

	return e.showBookableScheduleOptions(state, response)
}

func (e *Evaluator) showBookableScheduleOptions(state ChatState, response ChatResponse) (ChatResponse, error) {
	schedules, err := e.schedules.FindAllByDoctorID(state.SelectedDoctorID)
	if err != nil {
		return response, err
	}
	if len(schedules) == 0 {
		return e.replyWithState(response, state, state.SelectedDoctorName+" belum memiliki jadwal praktik.")
	}

	options := make([]ScheduleOption, 0, len(schedules))
	for _, schedule := range schedules {
		date, ok := nextDateForDay(schedule.DayOfWeek, time.Now())
		if !ok {
			continue
		}
		options = append(options, ScheduleOption{
			Number:       len(options) + 1,
			ScheduleID:   schedule.ID,
			DoctorID:     schedule.DoctorID,
			DoctorName:   state.SelectedDoctorName,
			Date:         date,
			DayOfWeek:    schedule.DayOfWeek,
			StartTime:    trimTime(schedule.StartTime),
			EndTime:      trimTime(schedule.EndTime),
			SlotInterval: schedule.SlotInterval,
		})
	}

	if len(options) == 0 {
		return e.replyWithState(response, state, "Jadwal dokter belum bisa dibaca karena format hari tidak dikenali.")
	}

	state.PendingSchedules = options
	state.PendingTimeSlots = nil
	state.Awaiting = awaitingScheduleSelection
	response.Data = options
	response.NeedInput = []string{awaitingScheduleSelection}
	response.Reply = "Pilih jadwal praktik:\n" + joinNumberedScheduleOptions(options) + "\n\nBalas dengan nomor jadwal."
	return e.finish(response, state)
}

func (e *Evaluator) selectScheduleByNumber(state ChatState, response ChatResponse, number int) (ChatResponse, error) {
	if number < 1 || number > len(state.PendingSchedules) {
		return e.replyWithState(response, state, fmt.Sprintf("Nomor jadwal tidak tersedia. Pilih nomor 1 sampai %d.", len(state.PendingSchedules)))
	}

	selected := state.PendingSchedules[number-1]
	state.SelectedDate = selected.Date

	booked, err := e.schedules.GetBookedAppointments(selected.DoctorID, selected.Date)
	if err != nil {
		return response, err
	}

	slots := buildTimeSlotOptions(selected, booked)
	if len(slots) == 0 {
		return e.replyWithState(response, state, "Tidak ada slot jam yang tersedia pada jadwal tersebut. Pilih jadwal lain:\n"+joinNumberedScheduleOptions(state.PendingSchedules))
	}

	state.PendingTimeSlots = slots
	state.Awaiting = awaitingTimeSelection
	response.Data = slots
	response.NeedInput = []string{awaitingTimeSelection}
	response.Reply = "Pilih jam booking:\n" + joinNumberedTimeSlotOptions(slots) + "\n\nBalas dengan nomor jam."
	return e.finish(response, state)
}

func (e *Evaluator) selectTimeByNumber(state ChatState, response ChatResponse, number int) (ChatResponse, error) {
	if number < 1 || number > len(state.PendingTimeSlots) {
		return e.replyWithState(response, state, fmt.Sprintf("Nomor jam tidak tersedia. Pilih nomor 1 sampai %d.", len(state.PendingTimeSlots)))
	}

	selected := state.PendingTimeSlots[number-1]
	if selected.Booked {
		return e.replyWithState(response, state, "Slot itu sudah terisi. Pilih jam lain:\n"+joinNumberedTimeSlotOptions(state.PendingTimeSlots))
	}

	state.SelectedTime = selected.Time
	state.Awaiting = awaitingPatientDetails
	response.NeedInput = []string{"name", "phone", "email"}
	response.Reply = "Masukkan data pasien dengan format:\nNama: Budi Santoso\nPhone: 081234567890\nEmail: budi@example.com"
	return e.finish(response, state)
}

func (e *Evaluator) createAppointmentFromPatientDetails(req ChatRequest, state ChatState, response ChatResponse) (ChatResponse, error) {
	patient := parsePatientDetails(req.Message)
	if !patient.Complete() {
		response.NeedInput = []string{"name", "phone", "email"}
		response.Reply = "Data pasien belum lengkap. Gunakan format:\nNama: Budi Santoso\nPhone: 081234567890\nEmail: budi@example.com"
		return e.finish(response, state)
	}

	state.PatientName = patient.Name
	state.PatientPhone = patient.Phone
	state.PatientEmail = patient.Email
	user, ok, err := e.resolveOrCreatePatient(req.ChatID, patient)
	if err != nil {
		return e.replyWithState(response, state, "Data pasien belum bisa disimpan: "+err.Error())
	}
	if !ok {
		return e.replyWithState(response, state, "Data pasien belum bisa disimpan. Coba gunakan email lain atau pastikan data pasien sudah benar.")
	}

	state.UserID = user.ID
	return e.createAppointmentFromState(req, state, response)
}

func (e *Evaluator) resolveOrCreatePatient(chatID string, patient patientDetails) (*models.User, bool, error) {
	user, err := e.users.FindByChatID(chatID)
	if err == nil {
		if existing, findErr := e.users.FindByEmail(patient.Email); findErr == nil && existing.ID != user.ID {
			applyPatientDetails(user, patient, false)
			if updateErr := e.users.Update(user); updateErr != nil {
				return nil, false, updateErr
			}
			return user, true, nil
		}

		applyPatientDetails(user, patient, true)
		if updateErr := e.users.Update(user); updateErr != nil {
			return nil, false, updateErr
		}
		return user, true, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, false, err
	}

	if existing, findErr := e.users.FindByEmail(patient.Email); findErr == nil {
		applyPatientDetails(existing, patient, false)
		if updateErr := e.users.Update(existing); updateErr != nil {
			return nil, false, updateErr
		}
		return existing, true, nil
	} else if findErr != gorm.ErrRecordNotFound {
		return nil, false, findErr
	}

	user = &models.User{
		ChatID:      chatID,
		FullName:    patient.Name,
		PhoneNumber: patient.Phone,
		Email:       patient.Email,
	}
	if createErr := e.users.Create(user); createErr != nil {
		return nil, false, createErr
	}
	return user, true, nil
}

func applyPatientDetails(user *models.User, patient patientDetails, includeEmail bool) {
	user.FullName = patient.Name
	user.PhoneNumber = patient.Phone
	if includeEmail {
		user.Email = patient.Email
	}
}

func (e *Evaluator) createAppointmentFromState(req ChatRequest, state ChatState, response ChatResponse) (ChatResponse, error) {
	appointmentDate, err := time.Parse("2006-01-02", state.SelectedDate)
	if err != nil {
		response.Reply = "Format tanggal belum valid. Silakan pilih jadwal lagi."
		state.Awaiting = awaitingScheduleSelection
		return e.finish(response, state)
	}

	appointment := models.Appointment{
		UserID:          state.UserID,
		DoctorID:        state.SelectedDoctorID,
		HospitalID:      state.SelectedHospitalID,
		AppointmentDate: appointmentDate,
		AppointmentTime: state.SelectedTime,
		Status:          models.StatusPending,
	}

	if err := e.appointments.Create(&appointment); err != nil {
		return e.replyWithState(response, state, "Booking belum bisa dibuat: "+err.Error())
	}

	e.stateStore.Clear(req.ChatID)
	response.Data = appointment
	response.State = nil
	patientName, patientPhone, patientEmail := state.PatientName, state.PatientPhone, state.PatientEmail
	if (patientName == "" || patientPhone == "" || patientEmail == "") && state.UserID != 0 {
		if user, findErr := e.users.FindByID(state.UserID); findErr == nil {
			patientName = user.FullName
			patientPhone = user.PhoneNumber
			patientEmail = user.Email
		}
	}
	response.Reply = bookingSuccessMessage(appointment, state, patientName, patientPhone, patientEmail)
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
