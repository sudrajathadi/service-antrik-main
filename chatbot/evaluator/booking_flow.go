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
	awaitingDate              = "DATE"
	awaitingTime              = "TIME"
	awaitingConfirmation      = "CONFIRMATION"
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
		state.SelectedDoctorID = doctor.ID
		state.SelectedDoctorName = doctor.Name
		state.SelectedHospitalID = doctor.HospitalID
		state.SelectedHospitalName = doctor.Hospital.Name
		state.SelectedSpecialty = doctor.Specialization.Name
	}

	missing := missingBookingFields(state)
	if len(missing) > 0 {
		if missing[0] == awaitingDoctor && state.SelectedSpecialty != "" {
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
				response.State = &state
				e.stateStore.Save(state)
				return response, nil
			}
		}
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

func (e *Evaluator) continueBookingFlow(req ChatRequest, state ChatState, response ChatResponse) (bool, ChatResponse, error) {
	if state.CurrentFlow != flowBooking {
		return false, response, nil
	}

	switch state.Awaiting {
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
		response.Reply = fmt.Sprintf("Nomor dokter tidak tersedia. Pilih nomor 1 sampai %d.", len(state.PendingDoctors))
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	selected := state.PendingDoctors[number-1]
	state.SelectedDoctorID = selected.ID
	state.SelectedDoctorName = selected.Name
	state.SelectedHospitalID = selected.HospitalID
	state.SelectedHospitalName = selected.Hospital
	state.SelectedSpecialty = selected.Specialization

	return e.showBookableScheduleOptions(state, response)
}

func (e *Evaluator) showBookableScheduleOptions(state ChatState, response ChatResponse) (ChatResponse, error) {
	schedules, err := e.schedules.FindAllByDoctorID(state.SelectedDoctorID)
	if err != nil {
		return response, err
	}
	if len(schedules) == 0 {
		response.Reply = state.SelectedDoctorName + " belum memiliki jadwal praktik."
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
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
		response.Reply = "Jadwal dokter belum bisa dibaca karena format hari tidak dikenali."
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	state.PendingSchedules = options
	state.PendingTimeSlots = nil
	state.Awaiting = awaitingScheduleSelection
	response.Data = options
	response.NeedInput = []string{awaitingScheduleSelection}
	response.Reply = "Pilih jadwal praktik:\n" + joinNumberedScheduleOptions(options) + "\n\nBalas dengan nomor jadwal."
	response.State = &state
	e.stateStore.Save(state)
	return response, nil
}

func (e *Evaluator) selectScheduleByNumber(state ChatState, response ChatResponse, number int) (ChatResponse, error) {
	if number < 1 || number > len(state.PendingSchedules) {
		response.Reply = fmt.Sprintf("Nomor jadwal tidak tersedia. Pilih nomor 1 sampai %d.", len(state.PendingSchedules))
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	selected := state.PendingSchedules[number-1]
	state.SelectedDate = selected.Date

	booked, err := e.schedules.GetBookedAppointments(selected.DoctorID, selected.Date)
	if err != nil {
		return response, err
	}

	slots := buildTimeSlotOptions(selected, booked)
	if len(slots) == 0 {
		response.Reply = "Tidak ada slot jam yang tersedia pada jadwal tersebut. Pilih jadwal lain:\n" + joinNumberedScheduleOptions(state.PendingSchedules)
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	state.PendingTimeSlots = slots
	state.Awaiting = awaitingTimeSelection
	response.Data = slots
	response.NeedInput = []string{awaitingTimeSelection}
	response.Reply = "Pilih jam booking:\n" + joinNumberedTimeSlotOptions(slots) + "\n\nBalas dengan nomor jam."
	response.State = &state
	e.stateStore.Save(state)
	return response, nil
}

func (e *Evaluator) selectTimeByNumber(state ChatState, response ChatResponse, number int) (ChatResponse, error) {
	if number < 1 || number > len(state.PendingTimeSlots) {
		response.Reply = fmt.Sprintf("Nomor jam tidak tersedia. Pilih nomor 1 sampai %d.", len(state.PendingTimeSlots))
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	selected := state.PendingTimeSlots[number-1]
	if selected.Booked {
		response.Reply = "Slot itu sudah terisi. Pilih jam lain:\n" + joinNumberedTimeSlotOptions(state.PendingTimeSlots)
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	state.SelectedTime = selected.Time
	state.Awaiting = awaitingPatientDetails
	response.NeedInput = []string{"name", "phone", "email"}
	response.Reply = "Masukkan data pasien dengan format:\nNama: Budi Santoso\nPhone: 081234567890\nEmail: budi@example.com"
	response.State = &state
	e.stateStore.Save(state)
	return response, nil
}

func (e *Evaluator) createAppointmentFromPatientDetails(req ChatRequest, state ChatState, response ChatResponse) (ChatResponse, error) {
	name, phone, email := parsePatientDetails(req.Message)
	if name == "" || phone == "" || email == "" {
		response.NeedInput = []string{"name", "phone", "email"}
		response.Reply = "Data pasien belum lengkap. Gunakan format:\nNama: Budi Santoso\nPhone: 081234567890\nEmail: budi@example.com"
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	state.PatientName = name
	state.PatientPhone = phone
	state.PatientEmail = email
	user, ok, err := e.resolveOrCreatePatient(req.ChatID, name, phone, email)
	if err != nil {
		response.Reply = "Data pasien belum bisa disimpan: " + err.Error()
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}
	if !ok {
		response.Reply = "Data pasien belum bisa disimpan. Coba gunakan email lain atau pastikan data pasien sudah benar."
		response.State = &state
		e.stateStore.Save(state)
		return response, nil
	}

	state.UserID = user.ID
	return e.createAppointmentFromState(req, state, response)
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

	return e.createAppointmentFromState(req, state, response)
}

func (e *Evaluator) resolveOrCreatePatient(chatID string, name string, phone string, email string) (*models.User, bool, error) {
	user, err := e.users.FindByChatID(chatID)
	if err == nil {
		if existing, findErr := e.users.FindByEmail(email); findErr == nil && existing.ID != user.ID {
			user.FullName = name
			user.PhoneNumber = phone
			if updateErr := e.users.Update(user); updateErr != nil {
				return nil, false, updateErr
			}
			return user, true, nil
		}

		user.FullName = name
		user.PhoneNumber = phone
		user.Email = email
		if updateErr := e.users.Update(user); updateErr != nil {
			return nil, false, updateErr
		}
		return user, true, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, false, err
	}

	if existing, findErr := e.users.FindByEmail(email); findErr == nil {
		existing.FullName = name
		existing.PhoneNumber = phone
		if updateErr := e.users.Update(existing); updateErr != nil {
			return nil, false, updateErr
		}
		return existing, true, nil
	} else if findErr != gorm.ErrRecordNotFound {
		return nil, false, findErr
	}

	user = &models.User{
		ChatID:      chatID,
		FullName:    name,
		PhoneNumber: phone,
		Email:       email,
	}
	if createErr := e.users.Create(user); createErr != nil {
		return nil, false, createErr
	}
	return user, true, nil
}

func (e *Evaluator) createAppointmentFromState(req ChatRequest, state ChatState, response ChatResponse) (ChatResponse, error) {
	appointmentDate, err := time.Parse("2006-01-02", state.SelectedDate)
	if err != nil {
		response.Reply = "Format tanggal belum valid. Silakan pilih jadwal lagi."
		state.Awaiting = awaitingScheduleSelection
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
