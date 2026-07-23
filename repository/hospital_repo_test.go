package repository

import "testing"

func TestHospitalNameSearchVariants(t *testing.T) {
	variants := hospitalNameSearchVariants("rumah sakit rs cengkareng")

	if !containsString(variants, "rs cengkareng") {
		t.Fatalf("expected variants to contain rs cengkareng, got %#v", variants)
	}
}

func TestHospitalNameSearchVariantsWithBranchNumber(t *testing.T) {
	variants := hospitalNameSearchVariants("primaya tangerang 2")

	if !containsString(variants, "primaya tangerang cabang 2") {
		t.Fatalf("expected variants to contain branch form, got %#v", variants)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
