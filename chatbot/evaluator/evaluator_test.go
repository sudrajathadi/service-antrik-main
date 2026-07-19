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

func TestBookingSuccessMessageIncludesPatientDetails(t *testing.T) {
	state := ChatState{
		SelectedDoctorName:   "drg. Galih Rahmawati",
		SelectedHospitalName: "RS Mitra Keluarga Bekasi",
		SelectedDate:         "2026-07-20",
		SelectedTime:         "08:00",
	}
	reply := bookingSuccessMessage(models.Appointment{Base: models.Base{ID: 7}}, state, "Sudrajat Hadi Susanto", "087775845951", "ajat5@gmail.com")

	for _, expected := range []string{
		"Nomor appointment: 7",
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
