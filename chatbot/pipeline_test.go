package chatbot

import (
	"encoding/json"
	"testing"
)

func TestTokenizeNormalizesSynonyms(t *testing.T) {
	tokens := Tokenize("Saya mau reservasi dr anak besok jam 10:00")

	if !containsToken(tokens, "booking") {
		t.Fatalf("expected reservasi to become booking, got %#v", tokens)
	}
	if !containsToken(tokens, "dokter") {
		t.Fatalf("expected dr to become dokter, got %#v", tokens)
	}
}

func TestParseBookingEntities(t *testing.T) {
	tokens := Tokenize("Saya mau booking dokter anak besok jam 10:00")
	parsed := Parse("Saya mau booking dokter anak besok jam 10:00", tokens)

	if parsed.Entities.Specialization != "anak" {
		t.Fatalf("expected anak specialization, got %q", parsed.Entities.Specialization)
	}
	if parsed.Entities.Date == "" {
		t.Fatal("expected parsed date for besok")
	}
	if parsed.Entities.Time != "10:00" {
		t.Fatalf("expected time 10:00, got %q", parsed.Entities.Time)
	}
}

func TestParseDoctorNameStopsBeforeDate(t *testing.T) {
	message := "jadwal dokter budi 2026-07-20"
	tokens := Tokenize(message)
	parsed := Parse(message, tokens)

	if parsed.Entities.DoctorName != "budi" {
		t.Fatalf("expected doctor name budi, got %q", parsed.Entities.DoctorName)
	}
	if parsed.Entities.Date != "2026-07-20" {
		t.Fatalf("expected date 2026-07-20, got %q", parsed.Entities.Date)
	}
}

func TestParseHospitalCityWithoutPreposition(t *testing.T) {
	tokens := Tokenize("rumah sakit tangerang ada apa saja?")
	parsed := Parse("rumah sakit tangerang ada apa saja?", tokens)

	if parsed.Entities.Location != "tangerang" {
		t.Fatalf("expected tangerang location, got %q", parsed.Entities.Location)
	}
}

func TestParseHospitalNameAndCityForDoctorQuestion(t *testing.T) {
	message := "rumah sakit bunda margonda depok ada dokter siapa saja?"
	tokens := Tokenize(message)
	parsed := Parse(message, tokens)

	if parsed.Entities.HospitalName != "bunda margonda" {
		t.Fatalf("expected hospital name bunda margonda, got %q", parsed.Entities.HospitalName)
	}
	if parsed.Entities.Location != "depok" {
		t.Fatalf("expected depok location, got %q", parsed.Entities.Location)
	}
}

func TestParseDoctorQuestionByHospitalDoesNotTreatLocationAsDoctorName(t *testing.T) {
	message := "Siapa dokter di RS Primaya Tangerang"
	tokens := Tokenize(message)
	parsed := Parse(message, tokens)

	if parsed.Entities.DoctorName != "" {
		t.Fatalf("expected empty doctor name, got %q", parsed.Entities.DoctorName)
	}
	if parsed.Entities.HospitalName != "primaya" {
		t.Fatalf("expected hospital name primaya, got %q", parsed.Entities.HospitalName)
	}
	if parsed.Entities.Location != "tangerang" {
		t.Fatalf("expected tangerang location, got %q", parsed.Entities.Location)
	}
}

func TestParseDoctorQuestionByHospitalWithoutPreposition(t *testing.T) {
	message := "siapa dokter rumah sakit cengkareng"
	tokens := Tokenize(message)
	parsed := Parse(message, tokens)

	if parsed.Entities.DoctorName != "" {
		t.Fatalf("expected empty doctor name, got %q", parsed.Entities.DoctorName)
	}
	if parsed.Entities.HospitalName != "cengkareng" {
		t.Fatalf("expected hospital name cengkareng, got %q", parsed.Entities.HospitalName)
	}
}

func TestParseBuildsSentenceTypeAndSyntaxTree(t *testing.T) {
	message := "jadwal dokter budi besok"
	tokens := Scan(message)
	parsed := Parse(message, tokens)

	if parsed.SentenceType != "PERINTAH" {
		t.Fatalf("expected PERINTAH sentence type, got %q", parsed.SentenceType)
	}
	if parsed.SyntaxTree.Type != "KALIMAT" {
		t.Fatalf("expected KALIMAT syntax tree root, got %q", parsed.SyntaxTree.Type)
	}
	if !syntaxTreeHasNode(parsed.SyntaxTree, "DOKTER", "budi") {
		t.Fatalf("expected syntax tree to contain doctor entity, got %#v", parsed.SyntaxTree)
	}
	if !syntaxTreeHasNode(parsed.SyntaxTree, "TANGGAL", parsed.Entities.Date) {
		t.Fatalf("expected syntax tree to contain date entity, got %#v", parsed.SyntaxTree)
	}
}

func TestTranslateCoreIntents(t *testing.T) {
	tests := []struct {
		name    string
		message string
		intent  Intent
	}{
		{name: "hospital list", message: "daftar rumah sakit", intent: IntentListHospitals},
		{name: "hospital list by city", message: "rumah sakit di tangerang ada apa saja?", intent: IntentListHospitals},
		{name: "doctors by hospital", message: "rumah sakit bunda margonda depok ada dokter siapa saja?", intent: IntentFindDoctorByHospital},
		{name: "doctors by hospital rs acronym", message: "Siapa dokter di RS Primaya Tangerang", intent: IntentFindDoctorByHospital},
		{name: "doctors by hospital branch number", message: "siapa dokter di RS Primaya Tangerang 2", intent: IntentFindDoctorByHospital},
		{name: "doctors by hospital with duplicate rs phrase", message: "siapa dokter di rumah sakit RS Cengkareng", intent: IntentFindDoctorByHospital},
		{name: "schedule doctors by hospital", message: "tolong jadwal dokter di rumah sakit RSUP Nasional", intent: IntentFindDoctorByHospital},
		{name: "hospital detail", message: "detail rumah sakit RSUP Nasional", intent: IntentAskHospitalLocation},
		{name: "specialization list", message: "list spesialisasi", intent: IntentListSpecializations},
		{name: "specialization list by hospital", message: "ada apa saja spesialisasi di rumah sakit bunda margonda depok", intent: IntentListSpecializationsByHospital},
		{name: "schedule with date", message: "jadwal dokter budi besok", intent: IntentAskDoctorSchedule},
		{name: "schedule hour question", message: "jam berapa jadwal dokter budi besok", intent: IntentAskDoctorSchedule},
		{name: "schedule when question", message: "kapan dokter budi besok", intent: IntentAskDoctorSchedule},
		{name: "schedule without date", message: "jadwal dokter budi", intent: IntentAskDoctorSchedule},
		{name: "booking", message: "booking dokter anak besok jam 10:00", intent: IntentBookAppointment},
		{name: "symptom is not recognized", message: "saya nyeri dada berat dan sulit bernapas", intent: IntentUnknown},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens := Tokenize(test.message)
			parsed := Parse(test.message, tokens)
			intent, _ := Translate(parsed)
			if intent != test.intent {
				t.Fatalf("expected %s, got %s", test.intent, intent)
			}
		})
	}
}

func syntaxTreeHasNode(node SyntaxNode, nodeType, value string) bool {
	if node.Type == nodeType && node.Value == value {
		return true
	}
	for _, child := range node.Children {
		if syntaxTreeHasNode(child, nodeType, value) {
			return true
		}
	}
	return false
}

func TestChatRequestAcceptsStringUserID(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		expected UserID
	}{
		{
			name:     "number",
			payload:  `{"chat_id":"chat-1","user_id":12,"message":"halo"}`,
			expected: 12,
		},
		{
			name:     "numeric string",
			payload:  `{"chat_id":"chat-1","user_id":"12","message":"halo"}`,
			expected: 12,
		},
		{
			name:     "phone string",
			payload:  `{"chat_id":"087775845951","user_id":"087775845951","message":"halo"}`,
			expected: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var request ChatRequest
			if err := json.Unmarshal([]byte(test.payload), &request); err != nil {
				t.Fatalf("expected payload to decode, got %v", err)
			}
			if request.UserID != test.expected {
				t.Fatalf("expected user_id %d, got %d", test.expected, request.UserID)
			}
		})
	}
}
