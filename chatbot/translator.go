package chatbot

func Translate(parsed ParseResult) (Intent, float64) {
	tokens := parsed.Tokens

	if parsed.IsNegation && containsToken(tokens, "batal", "cancel") {
		return IntentCancelFlow, 0.95
	}
	if parsed.IsConfirmation {
		return IntentConfirmBooking, 0.85
	}
	if hasEmergencyRedFlag(parsed) {
		return IntentEmergency, 0.95
	}
	if containsToken(tokens, "halo", "hai", "hello", "pagi", "siang", "sore", "malam") {
		return IntentGreeting, 0.8
	}
	if containsToken(tokens, "lokasi", "alamat") && containsToken(tokens, "rumah", "sakit") {
		return IntentAskHospitalLocation, 0.9
	}
	if isHospitalListQuestion(parsed) {
		return IntentListHospitals, 0.9
	}
	if containsPhrase(parsed.OriginalMessage, "list spesialisasi", "daftar spesialisasi") ||
		containsToken(tokens, "spesialisasi") && containsToken(tokens, "list", "daftar", "apa") {
		return IntentListSpecializations, 0.9
	}
	if containsToken(tokens, "jadwal") {
		return IntentAskDoctorSchedule, 0.9
	}
	if containsToken(tokens, "booking", "pesan") || containsPhrase(parsed.OriginalMessage, "buat janji", "janji temu") {
		return IntentBookAppointment, 0.92
	}
	if parsed.Entities.Specialization != "" && containsToken(tokens, "dokter") {
		return IntentFindDoctorBySpecialization, 0.88
	}
	if len(parsed.Entities.Symptoms) > 0 {
		return IntentRecommendSpecialization, 0.8
	}
	if containsToken(tokens, "dokter") {
		return IntentAskDoctor, 0.7
	}

	return IntentUnknown, 0.3
}

func isHospitalListQuestion(parsed ParseResult) bool {
	tokens := parsed.Tokens
	if !containsPhrase(parsed.OriginalMessage, "rumah sakit") {
		return false
	}
	return containsPhrase(parsed.OriginalMessage, "list rumah sakit", "daftar rumah sakit") ||
		containsToken(tokens, "list", "daftar", "tampilkan", "ada", "apa", "saja") ||
		parsed.Entities.Location != ""
}

func hasEmergencyRedFlag(parsed ParseResult) bool {
	return containsPhrase(
		parsed.OriginalMessage,
		"nyeri dada berat",
		"dada sesak",
		"sesak berat",
		"sulit bernapas",
		"tidak sadar",
		"pingsan",
		"kejang",
		"perdarahan hebat",
		"lumpuh mendadak",
		"bicara pelo",
	)
}
