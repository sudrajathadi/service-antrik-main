package chatbot

import (
	"regexp"
	"strings"
	"time"
)

var timePattern = regexp.MustCompile(`\b([01]?\d|2[0-3])[:.]([0-5]\d)\b`)
var datePattern = regexp.MustCompile(`\b(\d{4}-\d{2}-\d{2}|\d{1,2}/\d{1,2}/\d{4})\b`)
var selectionPattern = regexp.MustCompile(`^\d+$`)

// Parse mengubah token hasil tokenizer menjadi struktur entity yang bisa dipakai translator dan evaluator.
func Parse(message string, tokens []string) ParseResult {
	parsed := ParseResult{
		OriginalMessage: message,
		Tokens:          tokens,
		Entities:        Entities{},
		IsNegation:      containsToken(tokens, NegationTokens...),
	}

	for _, token := range parsed.Tokens {
		if isActionWord(token) {
			if !containsToken(parsed.ActionWords, token) {
				parsed.ActionWords = append(parsed.ActionWords, token)
			}
		}
		if spec, ok := SpecializationKeywordByToken[token]; ok {
			parsed.Entities.Specialization = spec
		}
	}

	parsed.Entities.DateText, parsed.Entities.Date = parseDateEntity(message)
	parsed.Entities.Time = parseTimeEntity(message)
	parsed.Entities.DoctorName = parseNamedEntityAfter(tokens, TokenDoctor)
	parsed.Entities.HospitalName = parseHospitalEntity(tokens)
	parsed.Entities.Location = parseLocationEntity(tokens)

	if parsed.Entities.HospitalName != "" {
		originalHospitalName := parsed.Entities.HospitalName
		hospitalName, city := splitHospitalNameAndCity(parsed.Entities.HospitalName)
		parsed.Entities.HospitalName = hospitalName
		if city != "" && (parsed.Entities.Location == "" || parsed.Entities.Location == originalHospitalName) {
			parsed.Entities.Location = city
		}
	}

	parsed.SentenceType = classifySentenceType(parsed)
	parsed.SyntaxTree = BuildSyntaxTree(parsed)

	return parsed
}

// parseTimeEntity membaca jam dari pesan dan menormalkannya ke format HH:MM.
// Contoh: "jam 9.30" menjadi "09:30".
func parseTimeEntity(message string) string {
	match := timePattern.FindStringSubmatch(strings.ReplaceAll(message, ".", ":"))
	if len(match) != 3 {
		return ""
	}
	hour := match[1]
	if len(hour) == 1 {
		hour = "0" + hour
	}
	return hour + ":" + match[2]
}

// parseDateEntity membaca tanggal relatif dan tanggal eksplisit, lalu menormalkannya ke format YYYY-MM-DD.
// Contoh: "besok" atau "20/07/2026" menjadi "2026-07-20".
func parseDateEntity(message string) (string, string) {
	today := time.Now()
	normalized := normalizeMessage(message)

	switch {
	case strings.Contains(normalized, TokenToday):
		return TokenToday, today.Format("2006-01-02")
	case strings.Contains(normalized, TokenTomorrow):
		return TokenTomorrow, today.AddDate(0, 0, 1).Format("2006-01-02")
	case strings.Contains(normalized, TokenDayAfter):
		return TokenDayAfter, today.AddDate(0, 0, 2).Format("2006-01-02")
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

// parseNamedEntityAfter mengambil nama setelah marker tertentu.
// Contoh marker "dokter" pada "jadwal dokter budi besok" menghasilkan "budi".
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

// parseHospitalEntity mengambil nama rumah sakit setelah frasa "rumah sakit".
// Contoh: "rumah sakit rsup nasional" menghasilkan "rsup nasional".
func parseHospitalEntity(tokens []string) string {
	for index := 0; index < len(tokens)-1; index++ {
		if tokens[index] == TokenHospitalFirst && tokens[index+1] == TokenHospitalSecond {
			return collectEntityWords(tokens[index+2:])
		}
	}
	return ""
}

// parseLocationEntity mengambil lokasi setelah marker lokasi seperti "di", "lokasi", atau "kota".
// Untuk pertanyaan list rumah sakit, lokasi juga bisa dibaca setelah frasa "rumah sakit".
func parseLocationEntity(tokens []string) string {
	for index, token := range tokens {
		if containsToken([]string{token}, LocationMarkerTokens...) {
			if index+1 < len(tokens) {
				if token == TokenIn && hasHospitalPhraseAt(tokens, index+1) {
					continue
				}
				return collectEntityWords(tokens[index+1:])
			}
		}
	}

	if containsToken(tokens, HospitalListIntentTokens...) {
		for index := 0; index < len(tokens)-1; index++ {
			if tokens[index] == TokenHospitalFirst && tokens[index+1] == TokenHospitalSecond {
				return collectEntityWords(tokens[index+2:])
			}
		}
	}

	return ""
}

func hasHospitalPhraseAt(tokens []string, index int) bool {
	return index+1 < len(tokens) &&
		tokens[index] == TokenHospitalFirst &&
		tokens[index+1] == TokenHospitalSecond
}

// splitHospitalNameAndCity memotong suffix kota yang dikenal dari nama rumah sakit.
// Contoh: "bunda margonda depok" menjadi "bunda margonda", "depok".
func splitHospitalNameAndCity(value string) (string, string) {
	tokens := strings.Fields(value)
	if len(tokens) < 2 {
		return value, ""
	}

	for _, city := range KnownCityPhrases {
		cityTokens := strings.Fields(city)
		if len(tokens) <= len(cityTokens) {
			continue
		}
		if hasSuffixTokens(tokens, cityTokens) {
			return strings.Join(tokens[:len(tokens)-len(cityTokens)], " "), city
		}
	}

	return value, ""
}

// hasSuffixTokens mengecek apakah token berakhir dengan rangkaian token tertentu.
func hasSuffixTokens(tokens []string, suffix []string) bool {
	if len(tokens) < len(suffix) {
		return false
	}
	offset := len(tokens) - len(suffix)
	for index := range suffix {
		if tokens[offset+index] != suffix[index] {
			return false
		}
	}
	return true
}

// collectEntityWords mengumpulkan kata entity sampai menemukan stopword, action word, tanggal, atau batas lain.
// Fungsi ini menjaga nama entity tidak ikut mengambil kata perintah atau tanggal.
func collectEntityWords(tokens []string) string {
	words := make([]string, 0, 3)
	for _, token := range tokens {
		if EntityStopwordTokens[token] || strings.Contains(token, ":") || datePattern.MatchString(token) {
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

// classifySentenceType memberi label tipe kalimat sederhana yang dipakai untuk menjelaskan konteks input.
func classifySentenceType(parsed ParseResult) string {
	normalized := normalizeMessage(parsed.OriginalMessage)
	hasPatientName := strings.Contains(normalized, "nama")
	hasPatientPhone := strings.Contains(normalized, "phone") || strings.Contains(normalized, "telepon") || strings.Contains(normalized, "telp")
	hasPatientEmail := strings.Contains(normalized, "email")

	switch {
	case hasPatientName && hasPatientPhone && hasPatientEmail:
		return "DATA_PASIEN"
	case len(parsed.Tokens) == 1 && selectionPattern.MatchString(parsed.Tokens[0]):
		return "PILIHAN_NOMOR"
	case parsed.IsNegation && containsToken(parsed.Tokens, TokenCancel, TokenCancelEN):
		return "PEMBATALAN"
	case containsToken(parsed.Tokens, GreetingTokens...):
		return "SAPAAN"
	case containsToken(parsed.Tokens, QuestionTokens...) || strings.Contains(parsed.OriginalMessage, "?"):
		return "PERTANYAAN"
	case containsToken(parsed.Tokens, BookingIntentTokens...) || containsPhrase(parsed.OriginalMessage, PhraseCreateAppointment, PhraseAppointmentMeeting):
		return "PERMINTAAN_BOOKING"
	case len(parsed.ActionWords) > 0:
		return "PERINTAH"
	default:
		return "PERNYATAAN"
	}
}

// BuildSyntaxTree menyusun pohon sintaks sederhana dari hasil scanner dan parser.
func BuildSyntaxTree(parsed ParseResult) SyntaxNode {
	return SyntaxNode{
		Type: "KALIMAT",
		Children: []SyntaxNode{
			{Type: "TIPE_KALIMAT", Value: parsed.SentenceType},
			{Type: "TOKEN", Children: buildValueNodes("KATA", parsed.Tokens)},
			{Type: "AKSI", Children: buildValueNodes("KATA_AKSI", parsed.ActionWords)},
			{Type: "ENTITY", Children: buildEntityNodes(parsed.Entities)},
			{Type: "KONTEKS", Children: buildContextNodes(parsed)},
		},
	}
}

func buildValueNodes(nodeType string, values []string) []SyntaxNode {
	nodes := make([]SyntaxNode, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		nodes = append(nodes, SyntaxNode{Type: nodeType, Value: value})
	}
	return nodes
}

func buildEntityNodes(entities Entities) []SyntaxNode {
	nodes := make([]SyntaxNode, 0, 6)
	appendEntityNode := func(name, value string) {
		if value != "" {
			nodes = append(nodes, SyntaxNode{Type: name, Value: value})
		}
	}

	appendEntityNode("SPESIALISASI", entities.Specialization)
	appendEntityNode("DOKTER", entities.DoctorName)
	appendEntityNode("RUMAH_SAKIT", entities.HospitalName)
	appendEntityNode("LOKASI", entities.Location)
	appendEntityNode("TANGGAL_TEXT", entities.DateText)
	appendEntityNode("TANGGAL", entities.Date)
	appendEntityNode("JAM", entities.Time)

	return nodes
}

func buildContextNodes(parsed ParseResult) []SyntaxNode {
	negation := "false"
	if parsed.IsNegation {
		negation = "true"
	}
	nodes := []SyntaxNode{{Type: "NEGASI", Value: negation}}
	if parsed.Entities.Date != "" {
		nodes = append(nodes, SyntaxNode{Type: "KONTEKS_WAKTU", Value: parsed.Entities.Date})
	}
	if parsed.Entities.Location != "" {
		nodes = append(nodes, SyntaxNode{Type: "KONTEKS_LOKASI", Value: parsed.Entities.Location})
	}
	return nodes
}
