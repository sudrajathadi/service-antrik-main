package evaluator

import (
	"strings"

	"service-antrik-chatbot/models"
)

func matchDoctor(doctors []models.Doctor, candidate string, message string) *models.Doctor {
	matches := matchDoctorCandidates(doctors, candidate, message)
	if len(matches) == 0 {
		return nil
	}
	return &matches[0]
}

func matchDoctorCandidates(doctors []models.Doctor, candidate string, message string) []models.Doctor {
	candidate = normalizeMessage(candidate)
	message = normalizeMessage(message)
	matches := make([]models.Doctor, 0)
	for index := range doctors {
		name := normalizeMessage(doctors[index].Name)
		if candidate != "" && strings.Contains(name, candidate) {
			matches = append(matches, doctors[index])
			continue
		}
		if strings.Contains(message, name) {
			matches = append(matches, doctors[index])
		}
	}
	return matches
}

func matchHospital(hospitals []models.Hospital, candidates ...string) *models.Hospital {
	for _, candidate := range candidates {
		candidate = normalizeMessage(candidate)
		if candidate == "" {
			continue
		}
		if hospital := matchHospitalBranchNumber(hospitals, candidate); hospital != nil {
			return hospital
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

func matchHospitalBranchNumber(hospitals []models.Hospital, candidate string) *models.Hospital {
	tokens := strings.Fields(candidate)
	if len(tokens) < 2 || !isDigitToken(tokens[len(tokens)-1]) {
		return nil
	}

	branchCandidate := strings.Join(append(tokens[:len(tokens)-1], "cabang", tokens[len(tokens)-1]), " ")
	branchCandidate = strings.ReplaceAll(branchCandidate, "rumah sakit", "rs")
	for index := range hospitals {
		name := strings.ReplaceAll(normalizeMessage(hospitals[index].Name), "rumah sakit", "rs")
		if strings.Contains(name, branchCandidate) {
			return &hospitals[index]
		}
	}
	return nil
}

func isDigitToken(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
