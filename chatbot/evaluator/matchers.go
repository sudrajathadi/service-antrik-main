package evaluator

import (
	"strings"

	"service-antrik-chatbot/models"
)

func matchDoctor(doctors []models.Doctor, candidate string, message string) *models.Doctor {
	candidate = normalizeMessage(candidate)
	message = normalizeMessage(message)
	for index := range doctors {
		name := normalizeMessage(doctors[index].Name)
		if candidate != "" && strings.Contains(name, candidate) {
			return &doctors[index]
		}
		if strings.Contains(message, name) {
			return &doctors[index]
		}
	}
	return nil
}

func matchHospital(hospitals []models.Hospital, candidates ...string) *models.Hospital {
	for _, candidate := range candidates {
		candidate = normalizeMessage(candidate)
		if candidate == "" {
			continue
		}
		for index := range hospitals {
			name := normalizeMessage(hospitals[index].Name)
			city := normalizeMessage(hospitals[index].City)
			if strings.Contains(name, candidate) || strings.Contains(candidate, name) || strings.Contains(city, candidate) {
				return &hospitals[index]
			}
		}
	}
	return nil
}
