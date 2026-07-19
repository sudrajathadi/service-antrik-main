package evaluator

import "service-antrik-chatbot/chatbot"

type Intent = chatbot.Intent
type ChatRequest = chatbot.ChatRequest
type ChatResponse = chatbot.ChatResponse
type ParseResult = chatbot.ParseResult
type ChatState = chatbot.ChatState
type StateStore = chatbot.StateStore
type DoctorSummary = chatbot.DoctorSummary
type ScheduleSummary = chatbot.ScheduleSummary
type ScheduleOption = chatbot.ScheduleOption
type TimeSlotOption = chatbot.TimeSlotOption

const (
	IntentGreeting                   = chatbot.IntentGreeting
	IntentCancelFlow                 = chatbot.IntentCancelFlow
	IntentListHospitals              = chatbot.IntentListHospitals
	IntentAskHospitalLocation        = chatbot.IntentAskHospitalLocation
	IntentListSpecializations        = chatbot.IntentListSpecializations
	IntentAskDoctor                  = chatbot.IntentAskDoctor
	IntentFindDoctorBySpecialization = chatbot.IntentFindDoctorBySpecialization
	IntentFindDoctorByHospital       = chatbot.IntentFindDoctorByHospital
	IntentAskDoctorSchedule          = chatbot.IntentAskDoctorSchedule
	IntentBookAppointment            = chatbot.IntentBookAppointment
)

func normalizeMessage(message string) string {
	return chatbot.NormalizeMessage(message)
}
