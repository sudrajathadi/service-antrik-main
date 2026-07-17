package chatbot

import (
	"regexp"
	"strings"
	"time"
)

var timePattern = regexp.MustCompile(`\b([01]?\d|2[0-3])[:.]([0-5]\d)\b`)
var datePattern = regexp.MustCompile(`\b(\d{4}-\d{2}-\d{2}|\d{1,2}/\d{1,2}/\d{4})\b`)

func Parse(message string, tokens []string) ParseResult {
	parsed := ParseResult{
		OriginalMessage: message,
		Tokens:          tokens,
		Entities:        Entities{},
		IsConfirmation:  containsToken(tokens, ConfirmationTokens...),
		IsNegation:      containsToken(tokens, NegationTokens...),
	}

	for _, token := range tokens {
		if isActionWord(token) {
			parsed.ActionWords = appendUnique(parsed.ActionWords, token)
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

	return parsed
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
		if tokens[index] == TokenHospitalFirst && tokens[index+1] == TokenHospitalSecond {
			return collectEntityWords(tokens[index+2:])
		}
	}
	return ""
}

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

func collectEntityWords(tokens []string) string {
	words := make([]string, 0, 3)
	for _, token := range tokens {
		if EntityStopwordTokens[token] || strings.Contains(token, ":") {
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
