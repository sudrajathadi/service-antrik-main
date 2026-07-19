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
	return e.finish(response, state)
}

func (e *Evaluator) hospitalLocation(state ChatState, response ChatResponse) (ChatResponse, error) {
	hospital, err := e.findHospitalFromParsed(response.Parsed)
	if err != nil {
		return response, err
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

func (e *Evaluator) findHospitalFromParsed(parsed ParseResult) (*models.Hospital, error) {
	hospitals, err := e.hospitals.FindAll()
	if err != nil {
		return nil, err
	}

	return matchHospital(hospitals, parsed.Entities.HospitalName, parsed.Entities.Location, parsed.OriginalMessage), nil
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

func (e *Evaluator) findDoctors(state ChatState, response ChatResponse) (ChatResponse, error) {
	filter := repository.DoctorFilter{
		Specialization: state.SelectedSpecialty,
		Location:       response.Parsed.Entities.Location,
		HospitalName:   response.Parsed.Entities.HospitalName,
	}

	if response.Parsed.Entities.HospitalName != "" {
		hospital, err := e.findHospitalFromParsed(response.Parsed)
		if err != nil {
			return response, err
		}
		if hospital == nil {
			response.Reply = "Saya belum menemukan rumah sakit " + response.Parsed.Entities.HospitalName + ". Coba cek nama rumah sakitnya."
			response.NeedInput = []string{"hospital_name"}
			return e.finish(response, state)
		}
		filter.HospitalID = hospital.ID
		filter.HospitalName = ""
		response.Parsed.Entities.HospitalName = hospital.Name
		response.Parsed.Entities.Location = hospital.City
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

func (e *Evaluator) showSchedule(state ChatState, response ChatResponse) (ChatResponse, error) {
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
