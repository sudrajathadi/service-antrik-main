package chatbot

func Translate(parsed ParseResult) (Intent, float64) {
	tokens := parsed.Tokens

	if parsed.IsNegation && containsToken(tokens, TokenCancel, TokenCancelEN) {
		return IntentCancelFlow, 0.95
	}
	if parsed.IsConfirmation {
		return IntentConfirmBooking, 0.85
	}
	if containsToken(tokens, GreetingTokens...) {
		return IntentGreeting, 0.8
	}
	if containsToken(tokens, HospitalLocationIntentTokens...) && containsToken(tokens, TokenHospitalFirst, TokenHospitalSecond) {
		return IntentAskHospitalLocation, 0.9
	}
	if containsPhrase(parsed.OriginalMessage, PhraseHospital) && containsToken(tokens, TokenDoctor) {
		return IntentFindDoctorByHospital, 0.92
	}
	if isHospitalListQuestion(parsed) {
		return IntentListHospitals, 0.9
	}
	if containsPhrase(parsed.OriginalMessage, PhraseListSpecialization, PhraseRegisterSpecialty) ||
		containsToken(tokens, TokenSpecialization) && containsToken(tokens, SpecializationListIntentTokens...) {
		return IntentListSpecializations, 0.9
	}
	if containsToken(tokens, TokenSchedule) {
		return IntentAskDoctorSchedule, 0.9
	}
	if containsToken(tokens, BookingIntentTokens...) || containsPhrase(parsed.OriginalMessage, PhraseCreateAppointment, PhraseAppointmentMeeting) {
		return IntentBookAppointment, 0.92
	}
	if parsed.Entities.Specialization != "" && containsToken(tokens, TokenDoctor) {
		return IntentFindDoctorBySpecialization, 0.88
	}
	if containsToken(tokens, TokenDoctor) {
		return IntentAskDoctor, 0.7
	}

	return IntentUnknown, 0.3
}

func isHospitalListQuestion(parsed ParseResult) bool {
	tokens := parsed.Tokens
	if !containsPhrase(parsed.OriginalMessage, PhraseHospital) {
		return false
	}
	return containsPhrase(parsed.OriginalMessage, PhraseListHospital, PhraseRegisterHospital) ||
		containsToken(tokens, HospitalListIntentTokens...) ||
		parsed.Entities.Location != ""
}
