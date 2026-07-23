package chatbot

import (
	"strings"
	"unicode"
)

const (
	// Token dasar yang dipakai parser dan translator.
	TokenDoctor         = "dokter"
	TokenHospitalFirst  = "rumah"
	TokenHospitalSecond = "sakit"
	TokenSchedule       = "jadwal"
	TokenBooking        = "booking"
	TokenLocation       = "lokasi"
	TokenAddress        = "alamat"
	TokenDetail         = "detail"
	TokenSpecialization = "spesialisasi"
	TokenList           = "list"
	TokenRegister       = "daftar"
	TokenShow           = "tampilkan"
	TokenExists         = "ada"
	TokenWhat           = "apa"
	TokenWho            = "siapa"
	TokenWhen           = "kapan"
	TokenHowMany        = "berapa"
	TokenHour           = "jam"
	TokenAll            = "saja"
	TokenCity           = "kota"
	TokenIn             = "di"
	TokenCancel         = "batal"
	TokenCancelEN       = "cancel"
	TokenToday          = "hari ini"
	TokenTomorrow       = "besok"
	TokenDayAfter       = "lusa"
)

const (
	// Frasa yang lebih mudah dicek dari teks asli setelah dinormalisasi.
	PhraseHospital           = "rumah sakit"
	PhraseListHospital       = "list rumah sakit"
	PhraseRegisterHospital   = "daftar rumah sakit"
	PhraseListSpecialization = "list spesialisasi"
	PhraseRegisterSpecialty  = "daftar spesialisasi"
	PhraseCreateAppointment  = "buat janji"
	PhraseAppointmentMeeting = "janji temu"
)

// TokenSynonyms berisi kamus normalisasi kata sebelum dipakai parser.
// Contoh:
// - "dr" menjadi "dokter"
// - "rs" menjadi "rumah sakit"
// - "reservasi" menjadi "booking"
// - "alamat" menjadi "lokasi"
var TokenSynonyms = map[string]string{
	"dr":          TokenDoctor,
	"doctor":      TokenDoctor,
	"rs":          PhraseHospital,
	"hospital":    PhraseHospital,
	"schedule":    TokenSchedule,
	"appointment": TokenBooking,
	"reservasi":   TokenBooking,
	"janji":       TokenBooking,
	"temu":        TokenBooking,
	TokenAddress:  TokenLocation,
	"info":        TokenDetail,
	"informasi":   TokenDetail,
	"profil":      TokenDetail,
	"dimana":      TokenLocation,
	TokenIn:       TokenIn,
}

var NegationTokens = []string{"tidak", "nggak", "ga", "gak", TokenCancel, TokenCancelEN, "salah", "bukan"}

var GreetingTokens = []string{"halo", "hai", "hello", "pagi", "siang", "sore", "malam"}

var HospitalLocationIntentTokens = []string{TokenLocation, TokenAddress, TokenDetail}

var HospitalListIntentTokens = []string{TokenList, TokenRegister, TokenShow, TokenExists, TokenWhat, TokenAll}

var SpecializationListIntentTokens = []string{TokenList, TokenRegister, TokenWhat}

var BookingIntentTokens = []string{TokenBooking, "pesan"}

var LocationMarkerTokens = []string{TokenIn, TokenLocation, TokenCity}

var QuestionTokens = []string{TokenWhat, TokenWho, TokenWhen, TokenHowMany, TokenLocation}

// ActionTokens dipakai parser untuk membedakan kata perintah dari nama entity.
// Contoh input: "tampilkan rumah sakit di depok"
// Output action_words: ["tampilkan"]
var ActionTokens = []string{
	"cari",
	TokenShow,
	"lihat",
	TokenList,
	TokenRegister,
	TokenSchedule,
	TokenLocation,
	TokenBooking,
	"pesan",
	"buat",
	"pilih",
	"mau",
	"butuh",
	"rekomendasi",
	"tanya",
}

// EntityStopwordTokens menghentikan pembacaan nama dokter/rumah sakit/kota.
// Contoh input: "jadwal dokter budi besok"
// Saat membaca nama dokter setelah token "dokter", parser berhenti di token "besok"
// sehingga output doctor_name menjadi "budi", bukan "budi besok".
var EntityStopwordTokens = map[string]bool{
	TokenExists:         true,
	"yang":              true,
	"untuk":             true,
	"hari":              true,
	"ini":               true,
	TokenTomorrow:       true,
	TokenDayAfter:       true,
	"jam":               true,
	"pukul":             true,
	TokenSchedule:       true,
	TokenBooking:        true,
	TokenLocation:       true,
	TokenIn:             true,
	TokenHospitalFirst:  true,
	TokenHospitalSecond: true,
	"dimana":            true,
	TokenWhat:           true,
}

// KnownCityPhrases membantu parser memisahkan "RS Bunda Margonda Depok"
// menjadi hospital_name = "bunda margonda" dan location = "depok".
var KnownCityPhrases = []string{
	"jakarta pusat",
	"jakarta selatan",
	"jakarta barat",
	"jakarta timur",
	"jakarta utara",
	"tangerang selatan",
	"tangerang",
	"bekasi",
	"depok",
	"bogor",
	"jakarta",
}

// SpecializationKeywordByToken membuat input seperti "dokter anak" menjadi spesialisasi "anak".
// Contoh input: "booking dokter gigi"
// Output entity specialization: "gigi"
var SpecializationKeywordByToken = map[string]string{
	"anak":         "anak",
	"pediatri":     "anak",
	"gigi":         "gigi",
	"mulut":        "gigi",
	"kulit":        "kulit",
	"kelamin":      "kulit",
	"jantung":      "jantung",
	"kardiologi":   "jantung",
	"mata":         "mata",
	"tht":          "tht",
	"telinga":      "tht",
	"hidung":       "tht",
	"tenggorokan":  "tht",
	"saraf":        "saraf",
	"neurologi":    "saraf",
	"kandungan":    "kandungan",
	"obgyn":        "kandungan",
	"ortopedi":     "ortopedi",
	"tulang":       "ortopedi",
	"penyakit":     "penyakit dalam",
	"dalam":        "penyakit dalam",
	"umum":         "umum",
	"paru":         "paru",
	"psikiater":    "jiwa",
	"jiwa":         "jiwa",
	"rehabilitasi": "rehabilitasi",
}

// Scan menjalankan proses scanner/tokenizer:
// 1. Normalisasi huruf dan tanda baca.
// 2. Memecah teks menjadi token.
// 3. Mengganti sinonim sesuai TokenSynonyms.
//
// Contoh:
// Input  : "Saya mau reservasi dr anak besok jam 10:00"
// Output : ["saya", "mau", "booking", "dokter", "anak", "besok", "jam", "10:00"]
//
// Contoh:
// Input  : "RS Bunda Margonda Depok"
// Output : ["rumah", "sakit", "bunda", "margonda", "depok"]
func Scan(message string) []string {
	normalized := normalizeMessage(message)
	parts := strings.Fields(normalized)
	tokens := make([]string, 0, len(parts))

	for _, part := range parts {
		if replacement, ok := TokenSynonyms[part]; ok {
			tokens = append(tokens, strings.Fields(replacement)...)
			continue
		}
		tokens = append(tokens, part)
	}

	return tokens
}

func Tokenize(message string) []string {
	return Scan(message)
}

// normalizeMessage membersihkan input sebelum dipecah menjadi token.
// Huruf dibuat kecil, spasi dirapikan, tanda baca umum diganti spasi,
// sedangkan ":", "-", dan "/" dipertahankan untuk format jam, telepon, atau tanggal.
//
// Contoh:
// Input  : "Detail RSUP Nasional, dong!"
// Output : "detail rsup nasional dong"
//
// Contoh:
// Input  : "Nama: Budi Phone: 0812-3456 Email: budi@mail.com"
// Output : "nama: budi phone: 0812-3456 email: budi mail com"
func normalizeMessage(message string) string {
	message = strings.ToLower(strings.TrimSpace(message))
	var builder strings.Builder
	builder.Grow(len(message))

	for _, r := range message {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			builder.WriteRune(r)
		case r == ':' || r == '-' || r == '/':
			builder.WriteRune(r)
		default:
			builder.WriteRune(' ')
		}
	}

	return strings.Join(strings.Fields(builder.String()), " ")
}

func NormalizeMessage(message string) string {
	return normalizeMessage(message)
}

func containsToken(tokens []string, values ...string) bool {
	for _, token := range tokens {
		for _, value := range values {
			if token == value {
				return true
			}
		}
	}
	return false
}

func containsPhrase(message string, values ...string) bool {
	normalized := normalizeMessage(message)
	for _, value := range values {
		if strings.Contains(normalized, normalizeMessage(value)) {
			return true
		}
	}
	return false
}

func isActionWord(token string) bool {
	return containsToken([]string{token}, ActionTokens...)
}
