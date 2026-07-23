package chatbot

type intentRule struct {
	Intent     Intent
	Confidence float64
	Match      func(ParseResult) bool
}

var intentRules = []intentRule{
	{IntentCancelFlow, 0.95, isCancelRequest},
	{IntentGreeting, 0.8, hasGreeting},
	{IntentListSpecializationsByHospital, 0.92, asksSpecializationsInHospital},
	{IntentAskHospitalLocation, 0.9, asksHospitalLocation},
	{IntentFindDoctorByHospital, 0.92, asksDoctorsInHospital},
	{IntentListHospitals, 0.9, isHospitalListQuestion},
	{IntentListSpecializations, 0.9, asksSpecializationList},
	{IntentAskDoctorSchedule, 0.9, hasScheduleToken},
	{IntentBookAppointment, 0.92, asksBooking},
	{IntentFindDoctorBySpecialization, 0.88, asksDoctorBySpecialization},
	{IntentAskDoctor, 0.7, hasDoctorToken},
}

func Translate(parsed ParseResult) (Intent, float64) {
	for _, rule := range intentRules {
		if rule.Match(parsed) {
			return rule.Intent, rule.Confidence
		}
	}
	return IntentUnknown, 0.3
}

func isHospitalListQuestion(parsed ParseResult) bool {
	tokens := parsed.Tokens
	if !hasHospitalReference(parsed) {
		return false
	}
	return containsPhrase(parsed.OriginalMessage, PhraseListHospital, PhraseRegisterHospital) ||
		containsToken(tokens, HospitalListIntentTokens...) ||
		parsed.Entities.Location != ""
}

func isCancelRequest(parsed ParseResult) bool {
	return parsed.IsNegation && containsToken(parsed.Tokens, TokenCancel, TokenCancelEN)
}

func hasGreeting(parsed ParseResult) bool {
	return containsToken(parsed.Tokens, GreetingTokens...)
}

func asksHospitalLocation(parsed ParseResult) bool {
	return containsToken(parsed.Tokens, HospitalLocationIntentTokens...) &&
		containsToken(parsed.Tokens, TokenHospitalFirst, TokenHospitalSecond)
}

func asksSpecializationsInHospital(parsed ParseResult) bool {
	return containsToken(parsed.Tokens, TokenSpecialization) &&
		hasHospitalReference(parsed)
}

func asksDoctorsInHospital(parsed ParseResult) bool {
	return hasHospitalReference(parsed) &&
		containsToken(parsed.Tokens, TokenDoctor)
}

func hasHospitalReference(parsed ParseResult) bool {
	return containsPhrase(parsed.OriginalMessage, PhraseHospital) ||
		containsToken(parsed.Tokens, TokenHospitalFirst) &&
			containsToken(parsed.Tokens, TokenHospitalSecond)
}

func asksSpecializationList(parsed ParseResult) bool {
	return containsPhrase(parsed.OriginalMessage, PhraseListSpecialization, PhraseRegisterSpecialty) ||
		containsToken(parsed.Tokens, TokenSpecialization) &&
			containsToken(parsed.Tokens, SpecializationListIntentTokens...)
}

func hasScheduleToken(parsed ParseResult) bool {
	return containsToken(parsed.Tokens, TokenSchedule, TokenWhen) ||
		containsPhrase(parsed.OriginalMessage, "jam berapa")
}

func asksBooking(parsed ParseResult) bool {
	return containsToken(parsed.Tokens, BookingIntentTokens...) ||
		containsPhrase(parsed.OriginalMessage, PhraseCreateAppointment, PhraseAppointmentMeeting)
}

func asksDoctorBySpecialization(parsed ParseResult) bool {
	return parsed.Entities.Specialization != "" && containsToken(parsed.Tokens, TokenDoctor)
}

func hasDoctorToken(parsed ParseResult) bool {
	return containsToken(parsed.Tokens, TokenDoctor)
}
