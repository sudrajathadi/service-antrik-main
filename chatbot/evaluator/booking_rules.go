package evaluator

func missingBookingFields(state ChatState) []string {
	var missing []string
	if state.SelectedDoctorID == 0 {
		missing = append(missing, awaitingDoctor)
	}
	if state.SelectedDate == "" {
		missing = append(missing, awaitingDate)
	}
	if state.SelectedTime == "" {
		missing = append(missing, awaitingTime)
	}
	if state.UserID == 0 && (state.PatientName == "" || state.PatientPhone == "" || state.PatientEmail == "") {
		missing = append(missing, awaitingPatientDetails)
	}
	return missing
}

func bookingQuestion(state ChatState, missing string) string {
	switch missing {
	case awaitingDoctor:
		if state.SelectedSpecialty != "" {
			return "Dokter untuk spesialisasi " + state.SelectedSpecialty + " yang mana? Sebutkan nama dokter yang dipilih."
		}
		return "Dokter atau spesialisasi apa yang ingin dibooking?"
	case awaitingDate:
		return "Untuk tanggal berapa? Kamu bisa tulis hari ini, besok, lusa, atau format YYYY-MM-DD."
	case awaitingTime:
		return "Jam berapa? Gunakan format HH:MM, contoh 10:00."
	case awaitingPatientDetails:
		return "Masukkan data pasien dengan format:\nNama: Budi Santoso\nPhone: 081234567890\nEmail: budi@example.com"
	default:
		return "Mohon lengkapi data booking."
	}
}
