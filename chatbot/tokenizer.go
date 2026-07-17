package chatbot

import (
	"strings"
	"unicode"
)

var tokenSynonyms = map[string]string{
	"dr":          "dokter",
	"doctor":      "dokter",
	"rs":          "rumah sakit",
	"hospital":    "rumah sakit",
	"schedule":    "jadwal",
	"appointment": "booking",
	"reservasi":   "booking",
	"janji":       "booking",
	"temu":        "booking",
	"alamat":      "lokasi",
	"dimana":      "lokasi",
	"di":          "di",
}

func Tokenize(message string) []string {
	normalized := normalizeMessage(message)
	parts := strings.Fields(normalized)
	tokens := make([]string, 0, len(parts))

	for _, part := range parts {
		if replacement, ok := tokenSynonyms[part]; ok {
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
