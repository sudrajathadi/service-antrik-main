package evaluator

import (
	"strings"
	"testing"

	"service-antrik-chatbot/models"
)

func TestParsePatientDetails(t *testing.T) {
	details := parsePatientDetails("Nama: Budi Santoso\nPhone: 0812-3456-7890\nEmail: budi@example.com")

	if details.Name != "Budi Santoso" {
		t.Fatalf("expected name Budi Santoso, got %q", details.Name)
	}
	if details.Phone != "081234567890" {
		t.Fatalf("expected normalized phone, got %q", details.Phone)
	}
	if details.Email != "budi@example.com" {
		t.Fatalf("expected email budi@example.com, got %q", details.Email)
	}
}

func TestHasPatientDetails(t *testing.T) {
	message := "Nama: Sudrajat Hadi Susanto\nPhone: 087775845951\nEmail: ajat5@gmail.com"

	if !hasPatientDetails(message) {
		t.Fatal("expected patient details message to be recognized")
	}
}

func TestParseSelectionNumber(t *testing.T) {
	number, ok := parseSelectionNumber("pilih nomor 2")
	if !ok || number != 2 {
		t.Fatalf("expected selection number 2, got %d ok=%v", number, ok)
	}
}

func TestMatchHospitalUnderstandsBranchNumber(t *testing.T) {
	hospitals := []models.Hospital{
		{Name: "RS Primaya Tangerang", City: "Tangerang"},
		{Name: "RS Primaya Tangerang Cabang 2 Tangerang", City: "Tangerang"},
	}

	hospital := matchHospital(hospitals, "primaya tangerang 2", "siapa dokter di rs primaya tangerang 2")
	if hospital == nil {
		t.Fatal("expected hospital to be matched")
	}
	if hospital.Name != "RS Primaya Tangerang Cabang 2 Tangerang" {
		t.Fatalf("expected branch hospital, got %q", hospital.Name)
	}
}

func TestShouldInterruptBookingFlowForNewHospitalDoctorQuestion(t *testing.T) {
	state := ChatState{CurrentFlow: flowBooking, Awaiting: awaitingComplaint}

	if !shouldInterruptBookingFlow(state, IntentFindDoctorByHospital) {
		t.Fatal("expected new hospital doctor question to interrupt active booking flow")
	}
}

func TestShouldNotInterruptBookingFlowForUnknownComplaint(t *testing.T) {
	state := ChatState{CurrentFlow: flowBooking, Awaiting: awaitingComplaint}

	if shouldInterruptBookingFlow(state, IntentUnknown) {
		t.Fatal("expected unknown complaint text to continue active booking flow")
	}
}

func TestBookingSuccessMessageIncludesPatientDetails(t *testing.T) {
	state := ChatState{
		SelectedDoctorName:   "drg. Galih Rahmawati",
		SelectedHospitalName: "RS Mitra Keluarga Bekasi",
		SelectedDate:         "2026-07-20",
		SelectedTime:         "08:00",
		PatientComplaint:     "sakit gigi sejak 2 hari",
	}
	reply := bookingSuccessMessage(models.Appointment{Base: models.Base{ID: 7}}, state, "Sudrajat Hadi Susanto", "087775845951", "ajat5@gmail.com")

	for _, expected := range []string{
		"Nomor appointment: 7",
		"Keluhan: sakit gigi sejak 2 hari",
		"Data pasien:",
		"Nama: Sudrajat Hadi Susanto",
		"Telepon: 087775845951",
		"Email: ajat5@gmail.com",
	} {
		if !strings.Contains(reply, expected) {
			t.Fatalf("expected reply to contain %q, got:\n%s", expected, reply)
		}
	}
}
