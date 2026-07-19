package chatbot

import (
	"regexp"
	"strings"
	"time"
)

var timePattern = regexp.MustCompile(`\b([01]?\d|2[0-3])[:.]([0-5]\d)\b`)
var datePattern = regexp.MustCompile(`\b(\d{4}-\d{2}-\d{2}|\d{1,2}/\d{1,2}/\d{4})\b`)

// Parse mengubah token hasil tokenizer menjadi struktur entity yang bisa dipakai translator dan evaluator.
func Parse(message string, tokens []string) ParseResult {
	parsed := newParseResult(message, tokens)
	readTokenEntities(&parsed)
	readPatternEntities(&parsed, message, tokens)
	normalizeHospitalAndCity(&parsed)

	return parsed
}

// newParseResult menyiapkan hasil parse dasar, termasuk flag negasi seperti "batal" atau "tidak".
func newParseResult(message string, tokens []string) ParseResult {
	return ParseResult{
		OriginalMessage: message,
		Tokens:          tokens,
		Entities:        Entities{},
		IsNegation:      containsToken(tokens, NegationTokens...),
	}
}

// readTokenEntities membaca entity yang bisa dikenali langsung dari satu token, seperti action word dan spesialisasi.
func readTokenEntities(parsed *ParseResult) {
	for _, token := range parsed.Tokens {
		if isActionWord(token) {
			parsed.ActionWords = appendUnique(parsed.ActionWords, token)
		}
		if spec, ok := SpecializationKeywordByToken[token]; ok {
			parsed.Entities.Specialization = spec
		}
	}
}

// readPatternEntities membaca entity yang perlu pola atau posisi kata, seperti tanggal, jam, dokter, rumah sakit, dan lokasi.
func readPatternEntities(parsed *ParseResult, message string, tokens []string) {
	parsed.Entities.DateText, parsed.Entities.Date = parseDateEntity(message)
	parsed.Entities.Time = parseTimeEntity(message)
	parsed.Entities.DoctorName = parseNamedEntityAfter(tokens, TokenDoctor)
	parsed.Entities.HospitalName = parseHospitalEntity(tokens)
	parsed.Entities.Location = parseLocationEntity(tokens)
}

// normalizeHospitalAndCity memisahkan nama rumah sakit dan kota jika user menulisnya dalam satu frasa.
// Contoh: "rumah sakit bunda margonda depok" menjadi hospital_name "bunda margonda" dan location "depok".
func normalizeHospitalAndCity(parsed *ParseResult) {
	if parsed.Entities.HospitalName == "" {
		return
	}

	originalHospitalName := parsed.Entities.HospitalName
	hospitalName, city := splitHospitalNameAndCity(parsed.Entities.HospitalName)
	parsed.Entities.HospitalName = hospitalName
	if city != "" && (parsed.Entities.Location == "" || parsed.Entities.Location == originalHospitalName) {
		parsed.Entities.Location = city
	}
}

// parseTimeEntity membaca jam dari pesan dan menormalkannya ke format HH:MM.
// Contoh: "jam 9.30" menjadi "09:30".
func parseTimeEntity(message string) string {
	match := timePattern.FindStringSubmatch(strings.ReplaceAll(message, ".", ":"))
	if len(match) != 3 {
		return ""
	}
	return leftPadHour(match[1]) + ":" + match[2]
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
				return collectEntityWords(tokens[index+1:])
			}
		}
	}

	if hasHospitalListWords(tokens) {
		for index := 0; index < len(tokens)-1; index++ {
			if tokens[index] == TokenHospitalFirst && tokens[index+1] == TokenHospitalSecond {
				return collectEntityWords(tokens[index+2:])
			}
		}
	}

	return ""
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
		if EntityStopwordTokens[token] || isEntityBoundaryToken(token) {
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

// isEntityBoundaryToken menandai token yang harus menghentikan pembacaan entity, misalnya jam atau tanggal.
func isEntityBoundaryToken(token string) bool {
	return strings.Contains(token, ":") || datePattern.MatchString(token)
}

// appendUnique menambahkan value ke slice hanya jika belum ada.
func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

// leftPadHour memastikan jam satu digit menjadi dua digit.
// Contoh: "9" menjadi "09".
func leftPadHour(value string) string {
	if len(value) == 1 {
		return "0" + value
	}
	return value
}
