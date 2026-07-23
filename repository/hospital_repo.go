package repository

import (
	"service-antrik-chatbot/models"
	"strings"

	"gorm.io/gorm"
)

type HospitalRepository interface {
	Create(hospital *models.Hospital) error
	FindAll() ([]models.Hospital, error)
	FindAllByCity(city string) ([]models.Hospital, error)
	FindAllByName(name string) ([]models.Hospital, error)
	FindByID(id uint) (*models.Hospital, error)
	Update(hospital *models.Hospital) error
	Delete(id uint) error
}

type hospitalRepository struct {
	db *gorm.DB
}

func NewHospitalRepository(db *gorm.DB) HospitalRepository {
	return &hospitalRepository{db}
}

func (r *hospitalRepository) Create(hospital *models.Hospital) error {
	return r.db.Create(hospital).Error
}

func (r *hospitalRepository) FindAll() ([]models.Hospital, error) {
	var hospitals []models.Hospital
	err := r.db.Find(&hospitals).Error
	return hospitals, err
}

func (r *hospitalRepository) FindAllByCity(city string) ([]models.Hospital, error) {
	city = strings.TrimSpace(city)
	if city == "" {
		return r.FindAll()
	}

	var hospitals []models.Hospital
	err := r.db.Where("city ILIKE ?", "%"+city+"%").Find(&hospitals).Error
	return hospitals, err
}

func (r *hospitalRepository) FindAllByName(name string) ([]models.Hospital, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return r.FindAll()
	}

	query := r.db.Model(&models.Hospital{})
	for index, variant := range hospitalNameSearchVariants(name) {
		condition := "name ILIKE ?"
		value := "%" + variant + "%"
		if index == 0 {
			query = query.Where(condition, value)
			continue
		}
		query = query.Or(condition, value)
	}

	var hospitals []models.Hospital
	err := query.Find(&hospitals).Error
	return hospitals, err
}

func (r *hospitalRepository) FindByID(id uint) (*models.Hospital, error) {
	var hospital models.Hospital
	err := r.db.First(&hospital, id).Error
	if err != nil {
		return nil, err
	}
	return &hospital, nil
}

func (r *hospitalRepository) Update(hospital *models.Hospital) error {
	return r.db.Save(hospital).Error
}

func (r *hospitalRepository) Delete(id uint) error {
	return r.db.Delete(&models.Hospital{}, id).Error
}

func hospitalNameSearchVariants(name string) []string {
	normalized := strings.ToLower(strings.Join(strings.Fields(name), " "))
	rawWithoutHospitalWords := strings.TrimSpace(strings.TrimPrefix(normalized, "rumah sakit "))
	rsVariant := normalized
	if rawWithoutHospitalWords != normalized {
		rsVariant = "rs " + rawWithoutHospitalWords
	}

	sourceVariants := []string{normalized, rawWithoutHospitalWords, rsVariant}
	if branchVariant := hospitalBranchSearchVariant(rawWithoutHospitalWords); branchVariant != "" {
		sourceVariants = append(sourceVariants, branchVariant)
	}

	seen := map[string]bool{}
	variants := make([]string, 0, 3)
	for _, variant := range sourceVariants {
		if variant == "" || seen[variant] {
			continue
		}
		seen[variant] = true
		variants = append(variants, variant)
	}
	return variants
}

func hospitalBranchSearchVariant(name string) string {
	tokens := strings.Fields(name)
	if len(tokens) < 2 || !isNumericString(tokens[len(tokens)-1]) {
		return ""
	}

	return strings.Join(append(tokens[:len(tokens)-1], "cabang", tokens[len(tokens)-1]), " ")
}

func isNumericString(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return value != ""
}
