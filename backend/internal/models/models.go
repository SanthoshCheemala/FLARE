package models

import "time"

type Customer struct {
	ID         int64     `json:"id"`
	ExternalID string    `json:"externalId"`
	Name       string    `json:"name"`
	DOB        string    `json:"dob"` // YYYY-MM-DD
	Country    string    `json:"country"`
	Hash       int64     `json:"hash"` // Changed to int64 for SQLite compatibility
	ListID     int64     `json:"listId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type CustomerList struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	FilePath    string    `json:"-"` // Internal use only
	RecordCount int       `json:"recordCount"`
	UploadedBy  int64     `json:"uploadedBy"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Sanction struct {
	ID        int64     `json:"id"`
	Source    string    `json:"source"` // OFAC, UN, EU
	Name      string    `json:"name"`
	DOB       string    `json:"dob"`
	Country   string    `json:"country"`
	Program   string    `json:"program"`
	Hash      int64     `json:"hash"` // Changed to int64 for SQLite compatibility
	ListID    int64     `json:"listId"`
	UpdatedAt time.Time `json:"updatedAt"`
	Version   int       `json:"version"`
}

type SanctionList struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Source      string    `json:"source"`
	Description string    `json:"description"`
	FilePath    string    `json:"-"` // Internal use only
	RecordCount int       `json:"recordCount"`
	Version     int       `json:"version"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Screening struct {
	ID               int64     `json:"id"`
	JobID            string    `json:"jobId"`
	Name             string    `json:"name"`
	CustomerListID   int64     `json:"customerListId"`
	SanctionListIDs  []int64   `json:"sanctionListIds"`
	Status           string    `json:"status"`
	MatchCount       int       `json:"matchCount"`
	CustomerCount    int       `json:"customerCount"`
	SanctionCount    int       `json:"sanctionCount"`
	WorkerCount      int       `json:"workerCount"`
	MemoryEstimateMB float64   `json:"memoryEstimateMb"`
	StartedAt        time.Time `json:"startedAt,omitempty"`
	FinishedAt       time.Time `json:"finishedAt,omitempty"`
	CreatedBy        int64     `json:"createdBy"`
	CreatedAt        time.Time `json:"createdAt"`
}

type ScreeningResult struct {
	ID             int64     `json:"id"`
	ScreeningID    int64     `json:"screeningId"`
	CustomerID     int64     `json:"customerId"`
	SanctionID     int64     `json:"sanctionId"`
	MatchScore     float64   `json:"matchScore"`
	Status         string    `json:"status"` // PENDING, CLEARED, FLAGGED
	InvestigatorID *int64    `json:"investigatorId,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ScreeningResultDetail struct {
	ScreeningResult
	Customer Customer `json:"customer"`
	Sanction Sanction `json:"sanction"`
}

type User struct {
	ID              int64      `json:"id"`
	Email           string     `json:"email"`
	PasswordHash    string     `json:"-"`
	Role            string     `json:"role"` // admin, compliance, viewer
	TwoFactorSecret string     `json:"-"`
	Active          bool       `json:"active"`
	LastLoginAt     *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type AuditLog struct {
	ID         int64                  `json:"id"`
	ActorID    int64                  `json:"actorId"`
	Action     string                 `json:"action"` // LOGIN, SCREENING_START, MATCH_UPDATE, etc.
	EntityType string                 `json:"entityType"`
	EntityID   string                 `json:"entityId"`
	Details    map[string]interface{} `json:"details,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}

type Settings struct {
	ID                int64     `json:"id"`
	FuzzyThreshold    float64   `json:"fuzzyThreshold"`
	DOBToleranceYears int       `json:"dobToleranceYears"`
	CountryMode       string    `json:"countryMode"` // EXACT, FUZZY
	RetentionYears    int       `json:"retentionYears"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// Request/Response DTOs

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TOTPCode string `json:"totpCode,omitempty"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	User         User   `json:"user"`
}

type StartScreeningRequest struct {
	Name            string            `json:"name"`
	CustomerListID  int64             `json:"customerListId"`
	SanctionListIDs []int64           `json:"sanctionListIds"`
	ColumnMapping   map[string]string `json:"columnMapping"`
}

type StartScreeningResponse struct {
	JobID string `json:"jobId"`
}

type UpdateMatchRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes,omitempty"`
}

// PSIDataPoint represents the standard data structure for PSI hashing
type PSIDataPoint struct {
	Name    string `json:"name"`
	DOB     string `json:"dob"`
	Country string `json:"country"`
}
