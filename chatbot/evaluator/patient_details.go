package evaluator

import (
	"regexp"
	"strings"
)

var (
	emailPattern = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	phonePattern = regexp.MustCompile(`(?:\+62|62|0)[0-9][0-9\-\s]{7,}`)
)

type patientDetails struct {
	Name  string
	Phone string
	Email string
}

func (details patientDetails) Complete() bool {
	return details.Name != "" && details.Phone != "" && details.Email != ""
}

func parsePatientDetails(message string) patientDetails {
	email := emailPattern.FindString(message)
	phone := normalizePhone(phonePattern.FindString(message))
	name := parseLabeledValue(message, "nama", "name")

	if phone == "" {
		phone = normalizePhone(parseLabeledValue(message, "phone", "telepon", "telp", "hp", "nomor"))
	}
	if email == "" {
		email = parseLabeledValue(message, "email", "mail")
	}
	if name == "" {
		name = inferNameFromPatientMessage(message, phone, email)
	}

	return patientDetails{
		Name:  strings.TrimSpace(name),
		Phone: strings.TrimSpace(phone),
		Email: strings.TrimSpace(email),
	}
}

func hasPatientDetails(message string) bool {
	return parsePatientDetails(message).Complete()
}

func parseLabeledValue(message string, labels ...string) string {
	lines := strings.FieldsFunc(message, func(r rune) bool {
		return r == '\n' || r == ',' || r == ';'
	})
	for _, line := range lines {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)
		for _, label := range labels {
			for _, prefix := range []string{label + ":", label + "="} {
				if strings.HasPrefix(lower, prefix) {
					return strings.TrimSpace(line[len(prefix):])
				}
			}
		}
	}
	return ""
}

func inferNameFromPatientMessage(message string, phone string, email string) string {
	cleaned := message
	if email != "" {
		cleaned = strings.ReplaceAll(cleaned, email, "")
	}
	if phone != "" {
		cleaned = strings.ReplaceAll(cleaned, phone, "")
	}
	cleaned = phonePattern.ReplaceAllString(cleaned, "")
	for _, word := range []string{"Nama:", "nama:", "Name:", "name:", "Phone:", "phone:", "Email:", "email:", "Telepon:", "telepon:"} {
		cleaned = strings.ReplaceAll(cleaned, word, "")
	}
	cleaned = strings.Trim(cleaned, " \n,;")
	return strings.Join(strings.Fields(cleaned), " ")
}

func normalizePhone(value string) string {
	value = strings.TrimSpace(value)
	replacer := strings.NewReplacer(" ", "", "-", "")
	return replacer.Replace(value)
}
