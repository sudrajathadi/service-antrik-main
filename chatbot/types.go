package chatbot

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"service-antrik-chatbot/models"
)

type Intent string

const (
	IntentUnknown                    Intent = "UNKNOWN"
	IntentGreeting                   Intent = "GREETING"
	IntentCancelFlow                 Intent = "CANCEL_FLOW"
	IntentConfirmBooking             Intent = "CONFIRM_BOOKING"
	IntentAskDoctor                  Intent = "ASK_DOCTOR"
	IntentAskDoctorSchedule          Intent = "ASK_DOCTOR_SCHEDULE"
	IntentListHospitals              Intent = "LIST_HOSPITALS"
	IntentAskHospitalLocation        Intent = "ASK_HOSPITAL_LOCATION"
	IntentListSpecializations        Intent = "LIST_SPECIALIZATIONS"
	IntentFindDoctorBySpecialization Intent = "FIND_DOCTOR_BY_SPECIALIZATION"
	IntentRecommendSpecialization    Intent = "RECOMMEND_SPECIALIZATION_BY_SYMPTOM"
	IntentBookAppointment            Intent = "BOOK_APPOINTMENT"
	IntentEmergency                  Intent = "EMERGENCY_RED_FLAG"
)

type ChatRequest struct {
	ChatID  string `json:"chat_id" binding:"required"`
	Message string `json:"message" binding:"required"`
	UserID  UserID `json:"user_id,omitempty"`
}

type UserID uint

func (id *UserID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*id = 0
		return nil
	}

	var number uint
	if err := json.Unmarshal(data, &number); err == nil {
		*id = UserID(number)
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	value = strings.TrimSpace(value)
	if value == "" || isPhoneLikeID(value) {
		*id = 0
		return nil
	}

	parsed, err := strconv.ParseUint(value, 10, 0)
	if err != nil {
		*id = 0
		return nil
	}

	*id = UserID(parsed)
	return nil
}

func isPhoneLikeID(value string) bool {
	return len(value) > 1 && strings.HasPrefix(value, "0")
}

type ChatResponse struct {
	ChatID     string      `json:"chat_id"`
	Intent     Intent      `json:"intent"`
	Reply      string      `json:"reply"`
	Tokens     []string    `json:"tokens,omitempty"`
	Parsed     ParseResult `json:"parsed"`
	State      *ChatState  `json:"state,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	NeedInput  []string    `json:"need_input,omitempty"`
	Confidence float64     `json:"confidence"`
}

type Token struct {
	Value string `json:"value"`
	Kind  string `json:"kind"`
}

type ParseResult struct {
	OriginalMessage string   `json:"original_message"`
	Tokens          []string `json:"tokens"`
	ActionWords     []string `json:"action_words,omitempty"`
	Entities        Entities `json:"entities"`
	IsConfirmation  bool     `json:"is_confirmation"`
	IsNegation      bool     `json:"is_negation"`
}

type Entities struct {
	Symptoms       []string `json:"symptoms,omitempty"`
	Specialization string   `json:"specialization,omitempty"`
	DoctorName     string   `json:"doctor_name,omitempty"`
	HospitalName   string   `json:"hospital_name,omitempty"`
	Location       string   `json:"location,omitempty"`
	DateText       string   `json:"date_text,omitempty"`
	Date           string   `json:"date,omitempty"`
	Time           string   `json:"time,omitempty"`
}

type ChatState struct {
	ChatID               string    `json:"chat_id"`
	CurrentFlow          string    `json:"current_flow,omitempty"`
	UserID               uint      `json:"user_id,omitempty"`
	SelectedDoctorID     uint      `json:"selected_doctor_id,omitempty"`
	SelectedDoctorName   string    `json:"selected_doctor_name,omitempty"`
	SelectedHospitalID   uint      `json:"selected_hospital_id,omitempty"`
	SelectedHospitalName string    `json:"selected_hospital_name,omitempty"`
	SelectedSpecialty    string    `json:"selected_specialization,omitempty"`
	SelectedDate         string    `json:"selected_date,omitempty"`
	SelectedTime         string    `json:"selected_time,omitempty"`
	SymptomsNote         string    `json:"symptoms_note,omitempty"`
	Awaiting             string    `json:"awaiting,omitempty"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type DoctorSummary struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Specialization string `json:"specialization"`
	HospitalID     uint   `json:"hospital_id"`
	Hospital       string `json:"hospital"`
	City           string `json:"city"`
	Experience     int    `json:"experience_years"`
}

type ScheduleSummary struct {
	DoctorID   uint              `json:"doctor_id"`
	DoctorName string            `json:"doctor_name"`
	DayOfWeek  string            `json:"day_of_week"`
	StartTime  string            `json:"start_time"`
	EndTime    string            `json:"end_time"`
	TimeSlots  []models.TimeSlot `json:"time_slots,omitempty"`
}
