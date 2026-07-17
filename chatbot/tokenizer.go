package chatbot

import (
	"strings"
	"unicode"
)

const (
	TokenDoctor         = "dokter"
	TokenHospitalFirst  = "rumah"
	TokenHospitalSecond = "sakit"
	TokenSchedule       = "jadwal"
	TokenBooking        = "booking"
	TokenLocation       = "lokasi"
	TokenAddress        = "alamat"
	TokenSpecialization = "spesialisasi"
	TokenList           = "list"
	TokenRegister       = "daftar"
	TokenShow           = "tampilkan"
	TokenExists         = "ada"
	TokenWhat           = "apa"
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
	PhraseHospital           = "rumah sakit"
	PhraseListHospital       = "list rumah sakit"
	PhraseRegisterHospital   = "daftar rumah sakit"
	PhraseListSpecialization = "list spesialisasi"
	PhraseRegisterSpecialty  = "daftar spesialisasi"
	PhraseCreateAppointment  = "buat janji"
	PhraseAppointmentMeeting = "janji temu"
)

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
	"dimana":      TokenLocation,
	TokenIn:       TokenIn,
}

var ConfirmationTokens = []string{"ya", "iya", "y", "ok", "oke", "setuju", "benar", "lanjut", "confirm", "konfirmasi"}

var NegationTokens = []string{"tidak", "nggak", "ga", "gak", TokenCancel, TokenCancelEN, "salah", "bukan"}

var GreetingTokens = []string{"halo", "hai", "hello", "pagi", "siang", "sore", "malam"}

var HospitalLocationIntentTokens = []string{TokenLocation, TokenAddress}

var HospitalListIntentTokens = []string{TokenList, TokenRegister, TokenShow, TokenExists, TokenWhat, TokenAll}

var SpecializationListIntentTokens = []string{TokenList, TokenRegister, TokenWhat}

var BookingIntentTokens = []string{TokenBooking, "pesan"}

var LocationMarkerTokens = []string{TokenIn, TokenLocation, TokenCity}

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

var EntityStopwordTokens = map[string]bool{
	TokenExists:   true,
	"yang":        true,
	"untuk":       true,
	"hari":        true,
	"ini":         true,
	TokenTomorrow: true,
	TokenDayAfter: true,
	"jam":         true,
	"pukul":       true,
	TokenSchedule: true,
	TokenBooking:  true,
	TokenLocation: true,
	"dimana":      true,
	TokenWhat:     true,
}

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

func Tokenize(message string) []string {
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

func hasHospitalListWords(tokens []string) bool {
	return containsToken(tokens, HospitalListIntentTokens...)
}
