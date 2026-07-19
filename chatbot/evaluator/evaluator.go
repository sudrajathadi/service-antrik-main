package evaluator

import "service-antrik-chatbot/repository"

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
	state := e.loadState(req)
	e.enrichStateFromParsed(&state, parsed)

	response := newResponse(req, parsed, intent, confidence)

	if intent != IntentCancelFlow {
		handled, flowResponse, err := e.continueBookingFlow(req, state, response)
		if handled || err != nil {
			return flowResponse, err
		}
	}

	if hasPatientDetails(req.Message) {
		response.Reply = "Data pasien sudah saya terima, tetapi belum ada booking aktif yang menunggu data pasien. Mulai dari pilih dokter, jadwal, dan jam terlebih dahulu, lalu kirim data pasien saat diminta."
		response.NeedInput = []string{"doctor", "schedule", "time"}
		return e.finish(response, state)
	}

	switch intent {
	case IntentGreeting:
		response.Reply = "Halo, saya bisa bantu cari dokter, cek jadwal, list rumah sakit, lokasi rumah sakit, spesialisasi, atau booking dokter."
	case IntentCancelFlow:
		e.stateStore.Clear(req.ChatID)
		response.Reply = "Baik, alur saat ini saya batalkan. Kamu bisa mulai lagi dengan tanya dokter, jadwal, rumah sakit, spesialisasi, atau booking."
		return response, nil
	case IntentListHospitals:
		return e.listHospitals(state, response)
	case IntentAskHospitalLocation:
		return e.hospitalLocation(state, response)
	case IntentListSpecializations:
		return e.listSpecializations(state, response)
	case IntentAskDoctor, IntentFindDoctorBySpecialization, IntentFindDoctorByHospital:
		return e.findDoctors(state, response)
	case IntentAskDoctorSchedule:
		return e.showSchedule(state, response)
	case IntentBookAppointment:
		return e.handleBooking(req, state, response)
	default:
		response.Reply = "Saya belum memahami pesan itu. Saya tidak menilai keluhan medis atau menentukan spesialisasi dari gejala. Kamu bisa bertanya seperti: list rumah sakit, lokasi rumah sakit, list spesialisasi, jadwal dokter, dokter anak, atau booking dokter."
	}

	return e.finish(response, state)
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
}

func (e *Evaluator) loadState(req ChatRequest) ChatState {
	state := e.stateStore.Get(req.ChatID)
	state.ChatID = req.ChatID
	if req.UserID != 0 {
		state.UserID = uint(req.UserID)
	}
	return state
}

func newResponse(req ChatRequest, parsed ParseResult, intent Intent, confidence float64) ChatResponse {
	return ChatResponse{
		ChatID:     req.ChatID,
		Intent:     intent,
		Tokens:     parsed.Tokens,
		Parsed:     parsed,
		Confidence: confidence,
	}
}

func (e *Evaluator) finish(response ChatResponse, state ChatState) (ChatResponse, error) {
	response.State = &state
	e.stateStore.Save(state)
	return response, nil
}

func (e *Evaluator) replyWithState(response ChatResponse, state ChatState, reply string) (ChatResponse, error) {
	response.Reply = reply
	return e.finish(response, state)
}
