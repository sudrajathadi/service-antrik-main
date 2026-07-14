package controllers

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"service-antrik-chatbot/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BulkUploadController struct {
	db *gorm.DB
}

type bulkUploadResult struct {
	Table        string   `json:"table"`
	InsertedRows int      `json:"inserted_rows"`
	Errors       []string `json:"errors,omitempty"`
}

type bulkUploadURLRequest struct {
	URL   string `json:"url" binding:"required"`
	GID   string `json:"gid"`
	Sheet string `json:"sheet"`
}

func NewBulkUploadController(db *gorm.DB) *BulkUploadController {
	return &BulkUploadController{db: db}
}

func (c *BulkUploadController) UploadCSV(ctx *gin.Context) {
	table := strings.TrimSpace(ctx.Param("table"))
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "csv file is required in multipart field 'file'"})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to open uploaded file: " + err.Error()})
		return
	}
	defer openedFile.Close()

	c.processCSVReader(ctx, table, openedFile)
}

func (c *BulkUploadController) UploadCSVFromURL(ctx *gin.Context) {
	table := strings.TrimSpace(ctx.Param("table"))

	var request bulkUploadURLRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	csvURL, err := normalizeSpreadsheetCSVURL(request.URL, request.GID, request.Sheet)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx.Request.Context(), http.MethodGet, csvURL, nil)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to build spreadsheet request: " + err.Error()})
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "failed to download spreadsheet csv: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  fmt.Sprintf("spreadsheet url is not publicly accessible as CSV. Upstream returned status %d", resp.StatusCode),
			"csvUrl": csvURL,
		})
		return
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(contentType, "text/html") {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  "spreadsheet url returned an HTML page, not CSV. Make sure the sheet is public or published to web",
			"csvUrl": csvURL,
		})
		return
	}

	c.processCSVReader(ctx, table, resp.Body)
}

func (c *BulkUploadController) processCSVReader(ctx *gin.Context, table string, csvInput io.Reader) {
	reader := csv.NewReader(csvInput)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read csv header: " + err.Error()})
		return
	}

	result, err := c.insertRows(ctx.Request.Context(), table, headers, reader)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(result.Errors) > 0 {
		ctx.JSON(http.StatusUnprocessableEntity, result)
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

func (c *BulkUploadController) DownloadTemplate(ctx *gin.Context) {
	table := strings.TrimSpace(ctx.Param("table"))
	fileName, ok := csvTemplateFiles[table]
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported table %q. supported tables: hospitals, specializations, doctors, doctor_schedules, users, appointments", table)})
		return
	}

	ctx.FileAttachment("csv_templates/"+fileName, fileName)
}

var csvTemplateFiles = map[string]string{
	"hospitals":        "hospitals.csv",
	"specializations":  "specializations.csv",
	"doctors":          "doctors.csv",
	"doctor_schedules": "doctor_schedules.csv",
	"users":            "users.csv",
	"appointments":     "appointments.csv",
}

func normalizeSpreadsheetCSVURL(rawURL string, requestedGID string, requestedSheet string) (string, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", errors.New("url must be a valid public http or https URL")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", errors.New("url must use http or https")
	}

	if strings.Contains(parsedURL.Host, "docs.google.com") && strings.HasPrefix(parsedURL.Path, "/spreadsheets/d/") {
		parts := strings.Split(parsedURL.Path, "/")
		if len(parts) < 4 || parts[3] == "" {
			return "", errors.New("google sheets url is missing spreadsheet id")
		}

		gid := strings.TrimSpace(requestedGID)
		if gid == "" {
			gid = parsedURL.Query().Get("gid")
		}
		if gid == "" && parsedURL.Fragment != "" {
			fragmentValues, _ := url.ParseQuery(parsedURL.Fragment)
			gid = fragmentValues.Get("gid")
		}

		if requestedSheet != "" {
			exportURL := url.URL{
				Scheme: "https",
				Host:   "docs.google.com",
				Path:   "/spreadsheets/d/" + parts[3] + "/gviz/tq",
			}
			query := exportURL.Query()
			query.Set("tqx", "out:csv")
			query.Set("sheet", requestedSheet)
			exportURL.RawQuery = query.Encode()

			return exportURL.String(), nil
		}

		if gid == "" {
			return "", errors.New("google sheets url does not include gid. Open the target tab and copy the URL with #gid=..., or send gid/sheet in the request payload")
		}

		exportURL := url.URL{
			Scheme: "https",
			Host:   "docs.google.com",
			Path:   "/spreadsheets/d/" + parts[3] + "/export",
		}
		query := exportURL.Query()
		query.Set("format", "csv")
		query.Set("gid", gid)
		exportURL.RawQuery = query.Encode()

		return exportURL.String(), nil
	}

	return parsedURL.String(), nil
}

func (c *BulkUploadController) insertRows(_ context.Context, table string, headers []string, reader *csv.Reader) (bulkUploadResult, error) {
	index := headerIndex(headers)
	result := bulkUploadResult{Table: table}

	switch table {
	case "hospitals":
		rows, errors := parseCSVRows(reader, func(rowNumber int, row []string) (models.Hospital, error) {
			return models.Hospital{
				Name:        requiredString(index, row, "name"),
				Address:     requiredString(index, row, "address"),
				City:        requiredString(index, row, "city"),
				PhoneNumber: optionalString(index, row, "phone_number"),
			}, validateRequired(index, row, rowNumber, "name", "address", "city")
		})
		result.Errors = errors
		if len(errors) == 0 && len(rows) > 0 {
			result.Errors = append(result.Errors, createBatch(c.db, rows)...)
			result.InsertedRows = len(rows)
		}
	case "specializations":
		rows, errors := parseCSVRows(reader, func(rowNumber int, row []string) (models.Specialization, error) {
			return models.Specialization{
				Name:        requiredString(index, row, "name"),
				Description: optionalString(index, row, "description"),
			}, validateRequired(index, row, rowNumber, "name")
		})
		result.Errors = errors
		if len(errors) == 0 && len(rows) > 0 {
			result.Errors = append(result.Errors, createBatch(c.db, rows)...)
			result.InsertedRows = len(rows)
		}
	case "doctors":
		rows, errors := parseCSVRows(reader, func(rowNumber int, row []string) (models.Doctor, error) {
			if err := validateRequired(index, row, rowNumber, "specialization_id", "hospital_id", "name"); err != nil {
				return models.Doctor{}, err
			}
			specializationID, err := requiredUint(index, row, rowNumber, "specialization_id")
			if err != nil {
				return models.Doctor{}, err
			}
			hospitalID, err := requiredUint(index, row, rowNumber, "hospital_id")
			if err != nil {
				return models.Doctor{}, err
			}
			experienceYears, err := optionalInt(index, row, rowNumber, "experience_years", 0)
			if err != nil {
				return models.Doctor{}, err
			}
			isActive, err := optionalBool(index, row, rowNumber, "is_active", true)
			if err != nil {
				return models.Doctor{}, err
			}
			return models.Doctor{
				SpecializationID: specializationID,
				HospitalID:       hospitalID,
				Name:             requiredString(index, row, "name"),
				Bio:              optionalString(index, row, "bio"),
				ExperienceYears:  experienceYears,
				IsActive:         isActive,
			}, nil
		})
		result.Errors = errors
		if len(errors) == 0 && len(rows) > 0 {
			result.Errors = append(result.Errors, createBatch(c.db, rows)...)
			result.InsertedRows = len(rows)
		}
	case "doctor_schedules":
		rows, errors := parseCSVRows(reader, func(rowNumber int, row []string) (models.DoctorSchedule, error) {
			if err := validateRequired(index, row, rowNumber, "doctor_id", "day_of_week", "start_time", "end_time"); err != nil {
				return models.DoctorSchedule{}, err
			}
			doctorID, err := requiredUint(index, row, rowNumber, "doctor_id")
			if err != nil {
				return models.DoctorSchedule{}, err
			}
			slotInterval, err := optionalInt(index, row, rowNumber, "slot_interval", 30)
			if err != nil {
				return models.DoctorSchedule{}, err
			}
			dayOfWeek, err := requiredDayOfWeek(index, row, rowNumber, "day_of_week")
			if err != nil {
				return models.DoctorSchedule{}, err
			}
			return models.DoctorSchedule{
				DoctorID:     doctorID,
				DayOfWeek:    dayOfWeek,
				StartTime:    requiredString(index, row, "start_time"),
				EndTime:      requiredString(index, row, "end_time"),
				SlotInterval: slotInterval,
			}, nil
		})
		result.Errors = errors
		if len(errors) == 0 && len(rows) > 0 {
			result.Errors = append(result.Errors, createBatch(c.db, rows)...)
			result.InsertedRows = len(rows)
		}
	case "users":
		rows, errors := parseCSVRows(reader, func(rowNumber int, row []string) (models.User, error) {
			return models.User{
				ChatID:      requiredString(index, row, "chat_id"),
				FullName:    requiredString(index, row, "full_name"),
				PhoneNumber: optionalString(index, row, "phone_number"),
				Email:       optionalString(index, row, "email"),
			}, validateRequired(index, row, rowNumber, "chat_id", "full_name")
		})
		result.Errors = errors
		if len(errors) == 0 && len(rows) > 0 {
			result.Errors = append(result.Errors, createBatch(c.db, rows)...)
			result.InsertedRows = len(rows)
		}
	case "appointments":
		rows, errors := parseCSVRows(reader, func(rowNumber int, row []string) (models.Appointment, error) {
			if err := validateRequired(index, row, rowNumber, "user_id", "doctor_id", "hospital_id", "appointment_date", "appointment_time"); err != nil {
				return models.Appointment{}, err
			}
			userID, err := requiredUint(index, row, rowNumber, "user_id")
			if err != nil {
				return models.Appointment{}, err
			}
			doctorID, err := requiredUint(index, row, rowNumber, "doctor_id")
			if err != nil {
				return models.Appointment{}, err
			}
			hospitalID, err := requiredUint(index, row, rowNumber, "hospital_id")
			if err != nil {
				return models.Appointment{}, err
			}
			appointmentDate, err := requiredDate(index, row, rowNumber, "appointment_date")
			if err != nil {
				return models.Appointment{}, err
			}
			status := models.AppointmentStatus(optionalString(index, row, "status"))
			if status == "" {
				status = models.StatusPending
			}
			return models.Appointment{
				UserID:          userID,
				DoctorID:        doctorID,
				HospitalID:      hospitalID,
				AppointmentDate: appointmentDate,
				AppointmentTime: requiredString(index, row, "appointment_time"),
				SymptomsNote:    optionalString(index, row, "symptoms_note"),
				Status:          status,
			}, nil
		})
		result.Errors = errors
		if len(errors) == 0 && len(rows) > 0 {
			result.Errors = append(result.Errors, createBatch(c.db, rows)...)
			result.InsertedRows = len(rows)
		}
	default:
		return result, fmt.Errorf("unsupported table %q. supported tables: hospitals, specializations, doctors, doctor_schedules, users, appointments", table)
	}

	if len(result.Errors) > 0 {
		result.InsertedRows = 0
	}

	return result, nil
}

func headerIndex(headers []string) map[string]int {
	index := make(map[string]int, len(headers))
	for i, header := range headers {
		index[strings.ToLower(strings.TrimSpace(header))] = i
	}
	return index
}

func parseCSVRows[T any](reader *csv.Reader, mapper func(rowNumber int, row []string) (T, error)) ([]T, []string) {
	var rows []T
	var errors []string
	rowNumber := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNumber++
		if err != nil {
			errors = append(errors, fmt.Sprintf("row %d: %v", rowNumber, err))
			continue
		}
		if isEmptyRow(record) {
			continue
		}

		row, err := mapper(rowNumber, record)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		rows = append(rows, row)
	}

	return rows, errors
}

func createBatch[T any](db *gorm.DB, rows []T) []string {
	if err := db.Create(&rows).Error; err != nil {
		return []string{err.Error()}
	}
	return nil
}

func validateRequired(index map[string]int, row []string, rowNumber int, fields ...string) error {
	for _, field := range fields {
		position, ok := index[field]
		if !ok {
			return fmt.Errorf("missing required header %q", field)
		}
		if position >= len(row) || strings.TrimSpace(row[position]) == "" {
			return fmt.Errorf("row %d: %s is required", rowNumber, field)
		}
	}
	return nil
}

func requiredString(index map[string]int, row []string, field string) string {
	return optionalString(index, row, field)
}

func requiredDayOfWeek(index map[string]int, row []string, rowNumber int, field string) (string, error) {
	value := strings.ToLower(optionalString(index, row, field))
	switch value {
	case "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday":
		return value, nil
	default:
		return "", fmt.Errorf("row %d: %s must be one of monday, tuesday, wednesday, thursday, friday, saturday, sunday", rowNumber, field)
	}
}

func optionalString(index map[string]int, row []string, field string) string {
	position, ok := index[field]
	if !ok || position >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[position])
}

func requiredUint(index map[string]int, row []string, rowNumber int, field string) (uint, error) {
	value := optionalString(index, row, field)
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("row %d: %s must be a positive number", rowNumber, field)
	}
	return uint(parsed), nil
}

func optionalInt(index map[string]int, row []string, rowNumber int, field string, fallback int) (int, error) {
	value := optionalString(index, row, field)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("row %d: %s must be a number", rowNumber, field)
	}
	return parsed, nil
}

func optionalBool(index map[string]int, row []string, rowNumber int, field string, fallback bool) (bool, error) {
	value := strings.ToLower(optionalString(index, row, field))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("row %d: %s must be true or false", rowNumber, field)
	}
	return parsed, nil
}

func requiredDate(index map[string]int, row []string, rowNumber int, field string) (time.Time, error) {
	value := optionalString(index, row, field)
	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("row %d: %s must use YYYY-MM-DD format", rowNumber, field)
}

func isEmptyRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
