package evaluator

import (
	"fmt"
	"strings"

	"service-antrik-chatbot/models"
	"service-antrik-chatbot/repository"
)

func (e *Evaluator) listHospitals(state ChatState, response ChatResponse) (ChatResponse, error) {
	city := strings.TrimSpace(response.Parsed.Entities.Location)
	var hospitals []models.Hospital
	var err error
	if city != "" {
		hospitals, err = e.hospitals.FindAllByCity(city)
	} else {
		hospitals, err = e.hospitals.FindAll()
	}
	if err != nil {
		return response, err
	}

	response.Data = hospitals
	if len(hospitals) == 0 {
		if city != "" {
			response.Reply = "Belum ada data rumah sakit yang tersedia untuk kota " + city + "."
		} else {
			response.Reply = "Belum ada data rumah sakit yang tersedia."
		}
	} else if city != "" {
		response.Reply = "Berikut daftar rumah sakit di " + city + ":\n\n" + joinHospitalNames(hospitals)
	} else {
		response.Reply = "Berikut daftar rumah sakit yang tersedia:\n\n" + joinHospitalNames(hospitals)
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
		HospitalName:   response.Parsed.Entities.HospitalName,
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
	if len(doctors) == 0 && response.Parsed.Entities.HospitalName != "" {
		response.Reply = "Saya belum menemukan dokter di rumah sakit " + response.Parsed.Entities.HospitalName + ". Coba cek nama rumah sakit atau kota yang dimaksud."
		response.NeedInput = []string{"hospital_name", "city"}
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
	state.CurrentFlow = flowBooking
	state.PendingDoctors = summaries
	state.PendingSchedules = nil
	state.PendingTimeSlots = nil
	state.Awaiting = awaitingDoctorSelection
	response.Data = summaries
	response.Reply = "Saya menemukan dokter berikut:\n" + joinNumberedDoctorNames(summaries) + "\n\nBalas dengan nomor dokter yang ingin dipilih."
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
