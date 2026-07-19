package evaluator

import "service-antrik-chatbot/models"

func rememberDoctor(state *ChatState, doctor models.Doctor) {
	state.SelectedDoctorID = doctor.ID
	state.SelectedDoctorName = doctor.Name
	state.SelectedHospitalID = doctor.HospitalID
	state.SelectedHospitalName = doctor.Hospital.Name
	state.SelectedSpecialty = doctor.Specialization.Name
}

func rememberDoctorSummary(state *ChatState, doctor DoctorSummary) {
	state.SelectedDoctorID = doctor.ID
	state.SelectedDoctorName = doctor.Name
	state.SelectedHospitalID = doctor.HospitalID
	state.SelectedHospitalName = doctor.Hospital
	state.SelectedSpecialty = doctor.Specialization
}
