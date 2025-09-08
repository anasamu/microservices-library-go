package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// StringUtils provides string utility functions
type StringUtils struct{}

// NewStringUtils creates a new StringUtils instance
func NewStringUtils() *StringUtils {
	return &StringUtils{}
}

// IsEmpty checks if a string is empty or contains only whitespace
func (su *StringUtils) IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty checks if a string is not empty
func (su *StringUtils) IsNotEmpty(s string) bool {
	return !su.IsEmpty(s)
}

// Truncate truncates a string to the specified length
func (su *StringUtils) Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// Capitalize capitalizes the first letter of a string
func (su *StringUtils) Capitalize(s string) string {
	if su.IsEmpty(s) {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// TitleCase converts a string to title case
func (su *StringUtils) TitleCase(s string) string {
	caser := cases.Title(language.English)
	return caser.String(s)
}

// Slugify converts a string to a URL-friendly slug
func (su *StringUtils) Slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove special characters
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	s = reg.ReplaceAllString(s, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")

	// Remove leading/trailing hyphens
	s = strings.Trim(s, "-")

	return s
}

// Contains checks if a string contains a substring (case-insensitive)
func (su *StringUtils) Contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// RandomString generates a random string of specified length
func (su *StringUtils) RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

// ValidationUtils provides validation utility functions
type ValidationUtils struct{}

// NewValidationUtils creates a new ValidationUtils instance
func NewValidationUtils() *ValidationUtils {
	return &ValidationUtils{}
}

// IsValidEmail validates an email address
func (vu *ValidationUtils) IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidPhone validates a phone number
func (vu *ValidationUtils) IsValidPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}

// IsValidURL validates a URL
func (vu *ValidationUtils) IsValidURL(url string) bool {
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return urlRegex.MatchString(url)
}

// IsValidUUID validates a UUID
func (vu *ValidationUtils) IsValidUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

// IsValidPassword validates a password (minimum 8 characters, at least one letter and one number)
func (vu *ValidationUtils) IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	return hasLetter && hasNumber
}

// CryptoUtils provides cryptographic utility functions
type CryptoUtils struct{}

// NewCryptoUtils creates a new CryptoUtils instance
func NewCryptoUtils() *CryptoUtils {
	return &CryptoUtils{}
}

// HashPassword hashes a password using bcrypt
func (cu *CryptoUtils) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against its hash
func (cu *CryptoUtils) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateHash generates a SHA256 hash of the input string
func (cu *CryptoUtils) GenerateHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// GenerateRandomBytes generates random bytes of specified length
func (cu *CryptoUtils) GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// TimeUtils provides time utility functions
type TimeUtils struct{}

// NewTimeUtils creates a new TimeUtils instance
func NewTimeUtils() *TimeUtils {
	return &TimeUtils{}
}

// Now returns the current time
func (tu *TimeUtils) Now() time.Time {
	return time.Now()
}

// NowUTC returns the current time in UTC
func (tu *TimeUtils) NowUTC() time.Time {
	return time.Now().UTC()
}

// FormatTime formats a time using the specified layout
func (tu *TimeUtils) FormatTime(t time.Time, layout string) string {
	return t.Format(layout)
}

// ParseTime parses a time string using the specified layout
func (tu *TimeUtils) ParseTime(timeStr, layout string) (time.Time, error) {
	return time.Parse(layout, timeStr)
}

// IsToday checks if a time is today
func (tu *TimeUtils) IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}

// IsYesterday checks if a time is yesterday
func (tu *TimeUtils) IsYesterday(t time.Time) bool {
	yesterday := time.Now().AddDate(0, 0, -1)
	return t.Year() == yesterday.Year() && t.YearDay() == yesterday.YearDay()
}

// IsThisWeek checks if a time is this week
func (tu *TimeUtils) IsThisWeek(t time.Time) bool {
	now := time.Now()
	year, week := now.ISOWeek()
	tYear, tWeek := t.ISOWeek()
	return year == tYear && week == tWeek
}

// IsThisMonth checks if a time is this month
func (tu *TimeUtils) IsThisMonth(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month()
}

// IsThisYear checks if a time is this year
func (tu *TimeUtils) IsThisYear(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year()
}

// AddDays adds days to a time
func (tu *TimeUtils) AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// AddMonths adds months to a time
func (tu *TimeUtils) AddMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// AddYears adds years to a time
func (tu *TimeUtils) AddYears(t time.Time, years int) time.Time {
	return t.AddDate(years, 0, 0)
}

// GetStartOfDay returns the start of the day for a given time
func (tu *TimeUtils) GetStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// GetEndOfDay returns the end of the day for a given time
func (tu *TimeUtils) GetEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// GetStartOfWeek returns the start of the week for a given time
func (tu *TimeUtils) GetStartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 0, but we want it to be 7
	}
	return tu.GetStartOfDay(t.AddDate(0, 0, -weekday+1))
}

// GetEndOfWeek returns the end of the week for a given time
func (tu *TimeUtils) GetEndOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 0, but we want it to be 7
	}
	return tu.GetEndOfDay(t.AddDate(0, 0, 7-weekday))
}

// GetStartOfMonth returns the start of the month for a given time
func (tu *TimeUtils) GetStartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// GetEndOfMonth returns the end of the month for a given time
func (tu *TimeUtils) GetEndOfMonth(t time.Time) time.Time {
	return tu.GetStartOfMonth(t.AddDate(0, 1, 0)).Add(-time.Nanosecond)
}

// GetStartOfYear returns the start of the year for a given time
func (tu *TimeUtils) GetStartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// GetEndOfYear returns the end of the year for a given time
func (tu *TimeUtils) GetEndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

// UUIDUtils provides UUID utility functions
type UUIDUtils struct{}

// NewUUIDUtils creates a new UUIDUtils instance
func NewUUIDUtils() *UUIDUtils {
	return &UUIDUtils{}
}

// Generate generates a new UUID
func (uu *UUIDUtils) Generate() uuid.UUID {
	return uuid.New()
}

// GenerateString generates a new UUID as a string
func (uu *UUIDUtils) GenerateString() string {
	return uuid.New().String()
}

// Parse parses a UUID string
func (uu *UUIDUtils) Parse(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}

// IsValid checks if a string is a valid UUID
func (uu *UUIDUtils) IsValid(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

// LogUtils provides logging utility functions
type LogUtils struct {
	logger *logrus.Logger
}

// NewLogUtils creates a new LogUtils instance
func NewLogUtils(logger *logrus.Logger) *LogUtils {
	return &LogUtils{logger: logger}
}

// LogError logs an error with context
func (lu *LogUtils) LogError(err error, message string, fields logrus.Fields) {
	lu.logger.WithError(err).WithFields(fields).Error(message)
}

// LogInfo logs an info message with context
func (lu *LogUtils) LogInfo(message string, fields logrus.Fields) {
	lu.logger.WithFields(fields).Info(message)
}

// LogWarning logs a warning message with context
func (lu *LogUtils) LogWarning(message string, fields logrus.Fields) {
	lu.logger.WithFields(fields).Warn(message)
}

// LogDebug logs a debug message with context
func (lu *LogUtils) LogDebug(message string, fields logrus.Fields) {
	lu.logger.WithFields(fields).Debug(message)
}

// LogWithDuration logs a message with duration
func (lu *LogUtils) LogWithDuration(message string, duration time.Duration, fields logrus.Fields) {
	fields["duration"] = duration
	lu.logger.WithFields(fields).Info(message)
}

// LogWithUser logs a message with user context
func (lu *LogUtils) LogWithUser(message string, userID, tenantID string, fields logrus.Fields) {
	fields["user_id"] = userID
	fields["tenant_id"] = tenantID
	lu.logger.WithFields(fields).Info(message)
}

// LogWithRequest logs a message with request context
func (lu *LogUtils) LogWithRequest(message string, requestID, method, path string, fields logrus.Fields) {
	fields["request_id"] = requestID
	fields["method"] = method
	fields["path"] = path
	lu.logger.WithFields(fields).Info(message)
}

// FileUtils provides file utility functions
type FileUtils struct{}

// NewFileUtils creates a new FileUtils instance
func NewFileUtils() *FileUtils {
	return &FileUtils{}
}

// GetFileExtension gets the file extension from a filename
func (fu *FileUtils) GetFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}
	return strings.ToLower(parts[len(parts)-1])
}

// GetFileNameWithoutExtension gets the filename without extension
func (fu *FileUtils) GetFileNameWithoutExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return filename
	}
	return strings.Join(parts[:len(parts)-1], ".")
}

// IsValidImageExtension checks if a file extension is a valid image extension
func (fu *FileUtils) IsValidImageExtension(extension string) bool {
	validExtensions := []string{"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg"}
	extension = strings.ToLower(extension)
	for _, validExt := range validExtensions {
		if extension == validExt {
			return true
		}
	}
	return false
}

// IsValidDocumentExtension checks if a file extension is a valid document extension
func (fu *FileUtils) IsValidDocumentExtension(extension string) bool {
	validExtensions := []string{"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "txt", "rtf"}
	extension = strings.ToLower(extension)
	for _, validExt := range validExtensions {
		if extension == validExt {
			return true
		}
	}
	return false
}

// FormatFileSize formats a file size in bytes to human readable format
func (fu *FileUtils) FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// Global utility functions for backward compatibility
var (
	StringUtils     = NewStringUtils()
	ValidationUtils = NewValidationUtils()
	CryptoUtils     = NewCryptoUtils()
	TimeUtils       = NewTimeUtils()
	UUIDUtils       = NewUUIDUtils()
	FileUtils       = NewFileUtils()
)

// Convenience functions
func IsEmpty(s string) bool {
	return StringUtils.IsEmpty(s)
}

func IsNotEmpty(s string) bool {
	return StringUtils.IsNotEmpty(s)
}

func IsValidEmail(email string) bool {
	return ValidationUtils.IsValidEmail(email)
}

func IsValidPhone(phone string) bool {
	return ValidationUtils.IsValidPhone(phone)
}

func IsValidURL(url string) bool {
	return ValidationUtils.IsValidURL(url)
}

func IsValidUUID(uuidStr string) bool {
	return ValidationUtils.IsValidUUID(uuidStr)
}

func IsValidPassword(password string) bool {
	return ValidationUtils.IsValidPassword(password)
}

func HashPassword(password string) (string, error) {
	return CryptoUtils.HashPassword(password)
}

func VerifyPassword(password, hash string) bool {
	return CryptoUtils.VerifyPassword(password, hash)
}

func GenerateHash(input string) string {
	return CryptoUtils.GenerateHash(input)
}

func Now() time.Time {
	return TimeUtils.Now()
}

func NowUTC() time.Time {
	return TimeUtils.NowUTC()
}

func GenerateUUID() uuid.UUID {
	return UUIDUtils.Generate()
}

func GenerateUUIDString() string {
	return UUIDUtils.GenerateString()
}
