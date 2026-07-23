package evaluator

import (
	"fmt"
	"sort"
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
	return e.finish(response, state)
}

func (e *Evaluator) hospitalLocation(state ChatState, response ChatResponse) (ChatResponse, error) {
	hospital, candidates, err := e.resolveHospitalFromParsed(response.Parsed)
	if err != nil {
		return response, err
	}
	if len(candidates) > 1 {
		return e.askHospitalSelection(state, response, candidates, IntentAskHospitalLocation)
	}

	if hospital == nil {
		hospitals, err := e.hospitals.FindAll()
		if err != nil {
			return response, err
		}
		response.Reply = "Rumah sakit yang mana? Sebutkan nama rumah sakitnya agar saya bisa tampilkan alamatnya."
		response.NeedInput = []string{"hospital_name"}
		response.Data = hospitals
		return e.finish(response, state)
	}

	response.Data = hospital
	response.Reply = fmt.Sprintf("%s berlokasi di %s, %s. Nomor telepon: %s.", hospital.Name, hospital.Address, hospital.City, emptyDash(hospital.PhoneNumber))
	return e.finish(response, state)
}

func (e *Evaluator) resolveHospitalFromParsed(parsed ParseResult) (*models.Hospital, []models.Hospital, error) {
	if parsed.Entities.HospitalName != "" {
		hospitals, err := e.hospitals.FindAllByName(parsed.Entities.HospitalName)
		if err != nil {
			return nil, nil, err
		}
		hospitals = filterHospitalsByLocation(hospitals, parsed.Entities.Location)
		if len(hospitals) > 1 {
			return nil, hospitals, nil
		}
		if len(hospitals) == 1 {
			return &hospitals[0], nil, nil
		}
	}

	if parsed.Entities.HospitalName == "" {
		return nil, nil, nil
	}

	hospitals, err := e.hospitals.FindAll()
	if err != nil {
		return nil, nil, err
	}

	return matchHospital(hospitals, parsed.Entities.HospitalName, parsed.Entities.Location, parsed.OriginalMessage), nil, nil
}

func filterHospitalsByLocation(hospitals []models.Hospital, location string) []models.Hospital {
	location = normalizeMessage(location)
	if location == "" {
		return hospitals
	}

	filtered := make([]models.Hospital, 0, len(hospitals))
	for _, hospital := range hospitals {
		name := normalizeMessage(hospital.Name)
		city := normalizeMessage(hospital.City)
		address := normalizeMessage(hospital.Address)
		if strings.Contains(name, location) || strings.Contains(city, location) || strings.Contains(address, location) {
			filtered = append(filtered, hospital)
		}
	}
	if len(filtered) == 0 {
		return hospitals
	}
	return filtered
}

func (e *Evaluator) askHospitalSelection(state ChatState, response ChatResponse, hospitals []models.Hospital, pendingIntent Intent) (ChatResponse, error) {
	summaries := summarizeHospitals(hospitals)
	state.CurrentFlow = ""
	state.Awaiting = awaitingHospitalSelection
	state.PendingIntent = pendingIntent
	state.PendingHospitals = summaries
	state.PendingDoctors = nil
	state.PendingSchedules = nil
	state.PendingTimeSlots = nil
	response.Data = summaries
	response.NeedInput = []string{awaitingHospitalSelection}
	response.Reply = "Saya menemukan beberapa rumah sakit yang cocok. Pilih rumah sakit yang dimaksud:\n" + joinNumberedHospitalNames(summaries) + "\n\nBalas dengan nomor rumah sakit."
	return e.finish(response, state)
}

func (e *Evaluator) selectHospitalByNumber(state ChatState, response ChatResponse, number int) (ChatResponse, error) {
	if number < 1 || number > len(state.PendingHospitals) {
		return e.replyWithState(response, state, fmt.Sprintf("Nomor rumah sakit tidak tersedia. Pilih nomor 1 sampai %d.", len(state.PendingHospitals)))
	}

	selected := state.PendingHospitals[number-1]
	hospital := models.Hospital{
		Base:        models.Base{ID: selected.ID},
		Name:        selected.Name,
		Address:     selected.Address,
		City:        selected.City,
		PhoneNumber: selected.PhoneNumber,
	}

	pendingIntent := state.PendingIntent
	state.SelectedHospitalID = hospital.ID
	state.SelectedHospitalName = hospital.Name
	state.PendingIntent = ""
	state.PendingHospitals = nil
	state.Awaiting = ""
	response.Parsed.Entities.HospitalName = hospital.Name
	response.Parsed.Entities.Location = hospital.City

	switch pendingIntent {
	case IntentAskHospitalLocation:
		response.Data = hospital
		response.Reply = fmt.Sprintf("%s berlokasi di %s, %s. Nomor telepon: %s.", hospital.Name, hospital.Address, hospital.City, emptyDash(hospital.PhoneNumber))
		return e.finish(response, state)
	case IntentListSpecializationsByHospital:
		return e.listSpecializationsForHospital(state, response, hospital)
	case IntentFindDoctorByHospital, IntentAskDoctor:
		return e.findDoctorsForHospital(state, response, hospital, repository.DoctorFilter{Specialization: state.SelectedSpecialty})
	default:
		response.Reply = "Pilihan rumah sakit sudah saya terima. Kamu ingin melihat lokasi, dokter, spesialisasi, atau jadwal?"
		response.NeedInput = []string{"intent"}
		return e.finish(response, state)
	}
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
	return e.finish(response, state)
}

func (e *Evaluator) listSpecializationsByHospital(state ChatState, response ChatResponse) (ChatResponse, error) {
	hospital, candidates, err := e.resolveHospitalFromParsed(response.Parsed)
	if err != nil {
		return response, err
	}
	if len(candidates) > 1 {
		return e.askHospitalSelection(state, response, candidates, IntentListSpecializationsByHospital)
	}
	if hospital == nil {
		response.Reply = "Rumah sakit yang mana? Sebutkan nama rumah sakitnya agar saya bisa tampilkan spesialisasi yang tersedia."
		response.NeedInput = []string{"hospital_name"}
		return e.finish(response, state)
	}

	return e.listSpecializationsForHospital(state, response, *hospital)
}

func (e *Evaluator) listSpecializationsForHospital(state ChatState, response ChatResponse, hospital models.Hospital) (ChatResponse, error) {
	doctors, err := e.doctors.FindAllFiltered(repository.DoctorFilter{HospitalID: hospital.ID})
	if err != nil {
		return response, err
	}

	seen := map[string]bool{}
	names := make([]string, 0)
	for _, doctor := range doctors {
		name := strings.TrimSpace(doctor.Specialization.Name)
		if !doctor.IsActive || name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	sort.Strings(names)

	response.Data = names
	response.Parsed.Entities.HospitalName = hospital.Name
	response.Parsed.Entities.Location = hospital.City
	if len(names) == 0 {
		response.Reply = "Belum ada data spesialisasi yang tersedia di " + hospital.Name + "."
	} else {
		lines := make([]string, 0, len(names))
		for _, name := range names {
			lines = append(lines, "- "+name)
		}
		response.Reply = "Spesialisasi yang tersedia di " + hospital.Name + ":\n" + strings.Join(lines, "\n")
	}
	return e.finish(response, state)
}

func (e *Evaluator) findDoctors(state ChatState, response ChatResponse) (ChatResponse, error) {
	filter := repository.DoctorFilter{
		Specialization: state.SelectedSpecialty,
		Location:       response.Parsed.Entities.Location,
		HospitalName:   response.Parsed.Entities.HospitalName,
	}

	if response.Parsed.Entities.HospitalName != "" {
		hospital, candidates, err := e.resolveHospitalFromParsed(response.Parsed)
		if err != nil {
			return response, err
		}
		if len(candidates) > 1 {
			return e.askHospitalSelection(state, response, candidates, IntentFindDoctorByHospital)
		}
		if hospital == nil {
			response.Reply = "Saya belum menemukan rumah sakit " + response.Parsed.Entities.HospitalName + ". Coba cek nama rumah sakitnya."
			response.NeedInput = []string{"hospital_name"}
			return e.finish(response, state)
		}
		return e.findDoctorsForHospital(state, response, *hospital, filter)
	}

	doctors, err := e.doctors.FindAllFiltered(filter)
	if err != nil {
		return response, err
	}

	if len(doctors) == 0 && state.SelectedSpecialty != "" {
		response.Reply = "Saya belum menemukan dokter untuk spesialisasi " + state.SelectedSpecialty + ". Coba pilih spesialisasi lain atau lihat daftar spesialisasi."
		response.NeedInput = []string{"specialization"}
		return e.finish(response, state)
	}
	if len(doctors) == 0 && response.Parsed.Entities.HospitalName != "" {
		response.Reply = "Saya belum menemukan dokter di rumah sakit " + response.Parsed.Entities.HospitalName + ". Coba cek nama rumah sakit atau kota yang dimaksud."
		response.NeedInput = []string{"hospital_name", "city"}
		return e.finish(response, state)
	}
	if len(doctors) == 0 {
		response.Reply = "Dokter yang mana atau spesialisasi apa yang ingin dicari?"
		response.NeedInput = []string{"doctor_name", "specialization"}
		return e.finish(response, state)
	}

	summaries := summarizeDoctors(doctors)
	state.CurrentFlow = flowBooking
	state.PendingDoctors = summaries
	state.PendingSchedules = nil
	state.PendingTimeSlots = nil
	state.Awaiting = awaitingDoctorSelection
	response.Data = summaries
	if response.Parsed.Entities.HospitalName != "" {
		response.Reply = "Saya menemukan dokter di " + response.Parsed.Entities.HospitalName + ":\n" + joinNumberedDoctorNames(summaries) + "\n\nBalas dengan nomor dokter yang ingin dipilih."
	} else {
		response.Reply = "Saya menemukan dokter berikut:\n" + joinNumberedDoctorNames(summaries) + "\n\nBalas dengan nomor dokter yang ingin dipilih."
	}
	return e.finish(response, state)
}

func (e *Evaluator) findDoctorsForHospital(state ChatState, response ChatResponse, hospital models.Hospital, filter repository.DoctorFilter) (ChatResponse, error) {
	filter.HospitalID = hospital.ID
	filter.HospitalName = ""
	filter.Location = ""
	response.Parsed.Entities.HospitalName = hospital.Name
	response.Parsed.Entities.Location = hospital.City

	doctors, err := e.doctors.FindAllFiltered(filter)
	if err != nil {
		return response, err
	}

	if len(doctors) == 0 {
		response.Reply = "Saya belum menemukan dokter di rumah sakit " + hospital.Name + ". Coba cek nama rumah sakit atau kota yang dimaksud."
		response.NeedInput = []string{"hospital_name", "city"}
		return e.finish(response, state)
	}

	summaries := summarizeDoctors(doctors)
	state.CurrentFlow = flowBooking
	state.PendingDoctors = summaries
	state.PendingHospitals = nil
	state.PendingSchedules = nil
	state.PendingTimeSlots = nil
	state.Awaiting = awaitingDoctorSelection
	response.Data = summaries
	response.Reply = "Saya menemukan dokter di " + hospital.Name + ":\n" + joinNumberedDoctorNames(summaries) + "\n\nBalas dengan nomor dokter yang ingin dipilih."
	return e.finish(response, state)
}

func (e *Evaluator) showSchedule(state ChatState, response ChatResponse) (ChatResponse, error) {
	if state.SelectedDoctorID == 0 && response.Parsed.Entities.DoctorName != "" {
		doctors, err := e.findDoctorCandidates(response.Parsed)
		if err != nil {
			return response, err
		}
		if len(doctors) > 1 {
			return e.askScheduleDoctorSelection(state, response, doctors)
		}
		if len(doctors) == 1 {
			rememberDoctor(&state, doctors[0])
		}
	}

	doctor, err := e.resolveDoctor(state, response.Parsed)
	if err != nil {
		return response, err
	}
	if doctor == nil {
		response.Reply = "Jadwal dokter siapa yang ingin dilihat? Sebutkan nama dokter dan tanggalnya, contoh: jadwal dokter Budi besok."
		response.NeedInput = []string{"doctor_name", "date"}
		return e.finish(response, state)
	}

	rememberDoctor(&state, *doctor)
	if state.SelectedDate == "" {
		response.Reply = "Untuk melihat slot yang sudah booked, sebutkan tanggalnya. Contoh: jadwal dokter " + doctor.Name + " besok."
		response.NeedInput = []string{"date"}
		return e.finish(response, state)
	}

	schedules, err := e.schedules.FindAllByDoctorID(doctor.ID)
	if err != nil {
		return response, err
	}

	summaries := make([]ScheduleSummary, 0, len(schedules))
	for _, schedule := range schedules {
		matchesDate, err := scheduleMatchesDate(schedule.DayOfWeek, state.SelectedDate)
		if err != nil {
			response.Reply = "Format tanggal belum valid. Gunakan format YYYY-MM-DD, atau tulis hari ini, besok, atau lusa."
			response.NeedInput = []string{"date"}
			return e.finish(response, state)
		}
		if !matchesDate {
			continue
		}

		booked, err := e.schedules.GetBookedAppointments(schedule.DoctorID, state.SelectedDate)
		if err != nil {
			return response, err
		}
		schedule.TimeSlots = markBookedSlots(booked, schedule.TimeSlots)
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
		response.Reply = doctor.Name + " tidak memiliki jadwal praktik pada tanggal " + state.SelectedDate + "."
	} else {
		response.Reply = fmt.Sprintf("Jadwal %s untuk tanggal %s:\n%s\n\nSlot dengan status booked sudah terisi.", doctor.Name, state.SelectedDate, joinSchedules(summaries))
	}

	return e.finish(response, state)
}

func (e *Evaluator) findDoctorCandidates(parsed ParseResult) ([]models.Doctor, error) {
	doctors, err := e.doctors.FindAll()
	if err != nil {
		return nil, err
	}
	return matchDoctorCandidates(doctors, parsed.Entities.DoctorName, parsed.OriginalMessage), nil
}

func (e *Evaluator) askScheduleDoctorSelection(state ChatState, response ChatResponse, doctors []models.Doctor) (ChatResponse, error) {
	summaries := summarizeDoctors(doctors)
	state.CurrentFlow = ""
	state.Awaiting = awaitingScheduleDoctor
	state.PendingIntent = IntentAskDoctorSchedule
	state.PendingDoctors = summaries
	state.PendingHospitals = nil
	state.PendingSchedules = nil
	state.PendingTimeSlots = nil
	response.Data = summaries
	response.NeedInput = []string{awaitingScheduleDoctor}
	response.Reply = "Saya menemukan beberapa dokter dengan nama tersebut. Pilih dokter yang dimaksud:\n" + joinNumberedDoctorNames(summaries) + "\n\nBalas dengan nomor dokter."
	return e.finish(response, state)
}

func (e *Evaluator) selectScheduleDoctorByNumber(state ChatState, response ChatResponse, number int) (ChatResponse, error) {
	if number < 1 || number > len(state.PendingDoctors) {
		return e.replyWithState(response, state, fmt.Sprintf("Nomor dokter tidak tersedia. Pilih nomor 1 sampai %d.", len(state.PendingDoctors)))
	}

	selected := state.PendingDoctors[number-1]
	state.SelectedDoctorID = selected.ID
	state.SelectedDoctorName = selected.Name
	state.SelectedHospitalID = selected.HospitalID
	state.SelectedHospitalName = selected.Hospital
	state.SelectedSpecialty = selected.Specialization
	state.PendingIntent = ""
	state.PendingDoctors = nil
	state.Awaiting = ""

	return e.showSchedule(state, response)
}
