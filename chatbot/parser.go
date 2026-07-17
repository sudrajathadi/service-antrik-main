package chatbot

import (
	"regexp"
	"strings"
	"time"
)

var timePattern = regexp.MustCompile(`\b([01]?\d|2[0-3])[:.]([0-5]\d)\b`)
var datePattern = regexp.MustCompile(`\b(\d{4}-\d{2}-\d{2}|\d{1,2}/\d{1,2}/\d{4})\b`)

var specializationKeywords = map[string]string{
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

var symptomKeywords = map[string]string{
	"demam":     "demam",
	"batuk":     "batuk",
	"pilek":     "pilek",
	"sesak":     "sesak napas",
	"napas":     "sesak napas",
	"dada":      "nyeri dada",
	"nyeri":     "nyeri",
	"sakit":     "sakit",
	"pusing":    "pusing",
	"kepala":    "sakit kepala",
	"gigi":      "sakit gigi",
	"gusi":      "sakit gigi",
	"perut":     "sakit perut",
	"mual":      "mual",
	"muntah":    "muntah",
	"diare":     "diare",
	"ruam":      "ruam kulit",
	"gatal":     "gatal",
	"mata":      "keluhan mata",
	"telinga":   "keluhan THT",
	"hamil":     "kehamilan",
	"haid":      "menstruasi",
	"sendi":     "nyeri sendi",
	"tulang":    "nyeri tulang",
	"kebas":     "kebas",
	"kesemutan": "kesemutan",
}

func Parse(message string, tokens []string) ParseResult {
	parsed := ParseResult{
		OriginalMessage: message,
		Tokens:          tokens,
		Entities:        Entities{},
		IsConfirmation:  containsToken(tokens, "ya", "iya", "y", "ok", "oke", "setuju", "benar", "lanjut", "confirm", "konfirmasi"),
		IsNegation:      containsToken(tokens, "tidak", "nggak", "ga", "gak", "batal", "cancel"),
	}

	for index, token := range tokens {
		if isActionWord(token) {
			parsed.ActionWords = appendUnique(parsed.ActionWords, token)
		}
		if spec, ok := specializationKeywords[token]; ok {
			parsed.Entities.Specialization = spec
		}
		if token == "sakit" && index > 0 && tokens[index-1] == "rumah" {
			continue
		}
		if symptom, ok := symptomKeywords[token]; ok {
			parsed.Entities.Symptoms = appendUnique(parsed.Entities.Symptoms, symptom)
		}
	}

	parsed.Entities.DateText, parsed.Entities.Date = parseDateEntity(message)
	parsed.Entities.Time = parseTimeEntity(message)
	parsed.Entities.DoctorName = parseNamedEntityAfter(tokens, "dokter")
	parsed.Entities.HospitalName = parseHospitalEntity(tokens)
	parsed.Entities.Location = parseLocationEntity(tokens)

	return parsed
}

func isActionWord(token string) bool {
	switch token {
	case "cari", "tampilkan", "lihat", "list", "daftar", "jadwal", "lokasi", "booking", "pesan", "buat", "pilih", "mau", "butuh", "rekomendasi", "tanya":
		return true
	default:
		return false
	}
}

func parseTimeEntity(message string) string {
	match := timePattern.FindStringSubmatch(strings.ReplaceAll(message, ".", ":"))
	if len(match) != 3 {
		return ""
	}
	return leftPadHour(match[1]) + ":" + match[2]
}

func parseDateEntity(message string) (string, string) {
	today := time.Now()
	normalized := normalizeMessage(message)

	switch {
	case strings.Contains(normalized, "hari ini"):
		return "hari ini", today.Format("2006-01-02")
	case strings.Contains(normalized, "besok"):
		return "besok", today.AddDate(0, 0, 1).Format("2006-01-02")
	case strings.Contains(normalized, "lusa"):
		return "lusa", today.AddDate(0, 0, 2).Format("2006-01-02")
	}

	match := datePattern.FindString(normalized)
	if match == "" {
		return "", ""
	}
	if parsed, err := time.Parse("2006-01-02", match); err == nil {
		return match, parsed.Format("2006-01-02")
	}
	if parsed, err := time.Parse("02/01/2006", match); err == nil {
		return match, parsed.Format("2006-01-02")
	}

	return match, ""
}

func parseNamedEntityAfter(tokens []string, marker string) string {
	for index, token := range tokens {
		if token != marker || index+1 >= len(tokens) {
			continue
		}

		candidate := collectEntityWords(tokens[index+1:])
		if candidate != "" && !isActionWord(candidate) {
			return candidate
		}
	}
	return ""
}

func parseHospitalEntity(tokens []string) string {
	for index := 0; index < len(tokens)-1; index++ {
		if tokens[index] == "rumah" && tokens[index+1] == "sakit" {
			return collectEntityWords(tokens[index+2:])
		}
	}
	return ""
}

func parseLocationEntity(tokens []string) string {
	for index, token := range tokens {
		if token == "di" || token == "lokasi" || token == "kota" {
			if index+1 < len(tokens) {
				return collectEntityWords(tokens[index+1:])
			}
		}
	}
	return ""
}

func collectEntityWords(tokens []string) string {
	stopWords := map[string]bool{
		"ada": true, "yang": true, "untuk": true, "hari": true, "ini": true,
		"besok": true, "lusa": true, "jam": true, "pukul": true, "jadwal": true,
		"booking": true, "lokasi": true, "dimana": true, "apa": true,
	}

	words := make([]string, 0, 3)
	for _, token := range tokens {
		if stopWords[token] || strings.Contains(token, ":") {
			break
		}
		if isActionWord(token) {
			break
		}
		words = append(words, token)
		if len(words) == 4 {
			break
		}
	}
	return strings.Join(words, " ")
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func leftPadHour(value string) string {
	if len(value) == 1 {
		return "0" + value
	}
	return value
}
