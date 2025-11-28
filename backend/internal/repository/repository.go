package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/SanthoshCheemala/FLARE/backend/internal/models"
	_ "github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Customer operations

func (r *Repository) CreateCustomerList(ctx context.Context, name, description, filePath string, uploadedBy int64) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO customer_lists (name, description, file_path, uploaded_by, created_at) 
		 VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		name, description, filePath, uploadedBy)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) CreateCustomer(ctx context.Context, c *models.Customer) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO customers (external_id, name, dob, country, hash, list_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		c.ExternalID, c.Name, c.DOB, c.Country, c.Hash, c.ListID)
	return err
}

func (r *Repository) GetCustomersByListID(ctx context.Context, listID int64) ([]models.Customer, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, external_id, name, dob, country, hash, list_id, created_at 
		 FROM customers WHERE list_id = ?`, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var c models.Customer
		if err := rows.Scan(&c.ID, &c.ExternalID, &c.Name, &c.DOB, &c.Country, &c.Hash, &c.ListID, &c.CreatedAt); err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}

	return customers, rows.Err()
}

func (r *Repository) GetCustomerSerializedStrings(ctx context.Context, listID int64) ([]string, error) {
	customers, err := r.GetCustomersByListID(ctx, listID)
	if err != nil {
		return nil, err
	}

	strings := make([]string, len(customers))
	for i, c := range customers {
		strings[i] = fmt.Sprintf("%s|%s|%s", c.Name, c.DOB, c.Country)
	}

	return strings, nil
}

func (r *Repository) GetCustomerLists(ctx context.Context) ([]models.CustomerList, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, description, file_path, record_count, uploaded_by, created_at FROM customer_lists ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []models.CustomerList
	for rows.Next() {
		var l models.CustomerList
		var filePath sql.NullString
		if err := rows.Scan(&l.ID, &l.Name, &l.Description, &filePath, &l.RecordCount, &l.UploadedBy, &l.CreatedAt); err != nil {
			return nil, err
		}
		if filePath.Valid {
			l.FilePath = filePath.String
		}
		lists = append(lists, l)
	}
	return lists, rows.Err()
}

// Sanction operations

func (r *Repository) CreateSanctionList(ctx context.Context, name, source, description, filePath string) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO sanction_lists (name, source, description, file_path, version, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		name, source, description, filePath)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) CreateSanction(ctx context.Context, s *models.Sanction) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO sanctions (source, name, dob, country, program, hash, list_id, updated_at, version)
		 VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, 1)`,
		s.Source, s.Name, s.DOB, s.Country, s.Program, s.Hash, s.ListID)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	s.ID = id
	return nil
}

func (r *Repository) GetSanctionsByListIDs(ctx context.Context, listIDs []int64) ([]models.Sanction, error) {
	if len(listIDs) == 0 {
		return []models.Sanction{}, nil
	}

	placeholders := make([]string, len(listIDs))
	args := make([]interface{}, len(listIDs))
	for i, id := range listIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`SELECT id, source, name, dob, country, program, hash, list_id, updated_at, version 
			  FROM sanctions WHERE list_id IN (%s)`, strings.Join(placeholders, ","))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sanctions []models.Sanction
	for rows.Next() {
		var s models.Sanction
		if err := rows.Scan(&s.ID, &s.Source, &s.Name, &s.DOB, &s.Country, &s.Program, &s.Hash, &s.ListID, &s.UpdatedAt, &s.Version); err != nil {
			return nil, err
		}
		sanctions = append(sanctions, s)
	}

	return sanctions, rows.Err()
}

func (r *Repository) GetSanctionSerializedStrings(ctx context.Context, listIDs []int64) ([]string, error) {
	sanctions, err := r.GetSanctionsByListIDs(ctx, listIDs)
	if err != nil {
		return nil, err
	}

	strings := make([]string, len(sanctions))
	for i, s := range sanctions {
		strings[i] = fmt.Sprintf("%s|%s|%s|%s", s.Name, s.DOB, s.Country, s.Program)
	}

	return strings, nil
}

func (r *Repository) GetSanctionLists(ctx context.Context) ([]models.SanctionList, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, source, description, file_path, record_count, version, updated_at, created_at FROM sanction_lists ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []models.SanctionList
	for rows.Next() {
		var l models.SanctionList
		var filePath sql.NullString
		if err := rows.Scan(&l.ID, &l.Name, &l.Source, &l.Description, &filePath, &l.RecordCount, &l.Version, &l.UpdatedAt, &l.CreatedAt); err != nil {
			return nil, err
		}
		if filePath.Valid {
			l.FilePath = filePath.String
		}
		lists = append(lists, l)
	}
	return lists, rows.Err()
}

// Screening operations

func (r *Repository) CreateScreening(ctx context.Context, s *models.Screening) error {
	// SanctionListIDs is stored as TEXT in schema; serialize slice -> comma-separated string
	var sanctionIDsStr string
	if len(s.SanctionListIDs) > 0 {
		parts := make([]string, len(s.SanctionListIDs))
		for i, id := range s.SanctionListIDs {
			parts[i] = fmt.Sprintf("%d", id)
		}
		sanctionIDsStr = strings.Join(parts, ",")
	} else {
		sanctionIDsStr = ""
	}

	res, err := r.db.ExecContext(ctx,
		`INSERT INTO screenings (job_id, name, customer_list_id, sanction_list_ids, status, 
		 customer_count, sanction_count, worker_count, memory_estimate_mb, created_by, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		s.JobID, s.Name, s.CustomerListID, sanctionIDsStr, s.Status,
		s.CustomerCount, s.SanctionCount, s.WorkerCount, s.MemoryEstimateMB, s.CreatedBy)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	s.ID = id
	return nil
}

func (r *Repository) UpdateScreeningStatus(ctx context.Context, jobID, status string, matchCount int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE screenings SET status = ?, match_count = ?, finished_at = CURRENT_TIMESTAMP WHERE job_id = ?`,
		status, matchCount, jobID)
	return err
}

// Screening result operations

func (r *Repository) CreateScreeningResult(ctx context.Context, sr *models.ScreeningResult) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO screening_results (screening_id, customer_id, sanction_id, match_score, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		sr.ScreeningID, sr.CustomerID, sr.SanctionID, sr.MatchScore, sr.Status)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	sr.ID = id
	return nil
}

// Audit log operations

func (r *Repository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_logs (actor_id, action, entity_type, entity_id, details, created_at)
		 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		log.ActorID, log.Action, log.EntityType, log.EntityID, log.Details)
	return err
}

func (r *Repository) GetScreeningResults(ctx context.Context, screeningID int64, limit, offset int) ([]models.ScreeningResultDetail, int, error) {
	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM screening_results WHERE screening_id = ?`, screeningID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results with joins
	rows, err := r.db.QueryContext(ctx,
		`SELECT sr.id, sr.screening_id, sr.customer_id, sr.sanction_id, sr.match_score, sr.status,
		        sr.investigator_id, sr.notes, sr.created_at, sr.updated_at,
		        c.id, c.external_id, c.name, c.dob, c.country, c.hash, c.list_id, c.created_at,
		        s.id, s.source, s.name, s.dob, s.country, s.program, s.hash, s.list_id, s.updated_at, s.version
		 FROM screening_results sr
		 JOIN customers c ON sr.customer_id = c.id
		 JOIN sanctions s ON sr.sanction_id = s.id
		 WHERE sr.screening_id = ?
		 ORDER BY sr.match_score DESC, sr.created_at DESC
		 LIMIT ? OFFSET ?`,
		screeningID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []models.ScreeningResultDetail
	for rows.Next() {
		var r models.ScreeningResultDetail
		err := rows.Scan(
			&r.ID, &r.ScreeningID, &r.CustomerID, &r.SanctionID, &r.MatchScore, &r.Status,
			&r.InvestigatorID, &r.Notes, &r.CreatedAt, &r.UpdatedAt,
			&r.Customer.ID, &r.Customer.ExternalID, &r.Customer.Name, &r.Customer.DOB,
			&r.Customer.Country, &r.Customer.Hash, &r.Customer.ListID, &r.Customer.CreatedAt,
			&r.Sanction.ID, &r.Sanction.Source, &r.Sanction.Name, &r.Sanction.DOB,
			&r.Sanction.Country, &r.Sanction.Program, &r.Sanction.Hash, &r.Sanction.ListID,
			&r.Sanction.UpdatedAt, &r.Sanction.Version,
		)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, r)
	}

	return results, total, rows.Err()
}

//  GetScreeningResultsByJobID gets results by job_id instead of screening_id
func (r *Repository) GetScreeningResultsByJobID(ctx context.Context, jobID string, limit, offset int) ([]models.ScreeningResultDetail, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT sr.id, sr.screening_id, sr.customer_id, sr.sanction_id, sr.match_score, sr.status,
		        sr.investigator_id, COALESCE(sr.notes, ''), sr.created_at, sr.updated_at,
		        c.id, c.external_id, c.name, c.dob, c.country, c.hash, c.list_id, c.created_at,
		        s.id, s.source, s.name, s.dob, s.country, s.program, s.hash, s.list_id, s.updated_at, s.version
		 FROM screening_results sr
		 JOIN screenings sc ON sr.screening_id = sc.id
		 JOIN customers c ON sr.customer_id = c.id
		 JOIN sanctions s ON sr.sanction_id = s.id
		 WHERE sc.job_id = ?
		 ORDER BY sr.match_score DESC, sr.created_at DESC
		 LIMIT ? OFFSET ?`,
		jobID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.ScreeningResultDetail
	for rows.Next() {
		var r models.ScreeningResultDetail
		err := rows.Scan(
			&r.ID, &r.ScreeningID, &r.CustomerID, &r.SanctionID, &r.MatchScore, &r.Status,
			&r.InvestigatorID, &r.Notes, &r.CreatedAt, &r.UpdatedAt,
			&r.Customer.ID, &r.Customer.ExternalID, &r.Customer.Name, &r.Customer.DOB,
			&r.Customer.Country, &r.Customer.Hash, &r.Customer.ListID, &r.Customer.CreatedAt,
			&r.Sanction.ID, &r.Sanction.Source, &r.Sanction.Name, &r.Sanction.DOB,
			&r.Sanction.Country, &r.Sanction.Program, &r.Sanction.Hash, &r.Sanction.ListID,
			&r.Sanction.UpdatedAt, &r.Sanction.Version,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, rows.Err()
}

// CountScreeningResultsByJobID counts results for a job
func (r *Repository) CountScreeningResultsByJobID(ctx context.Context, jobID string) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM screening_results sr
		 JOIN screenings sc ON sr.screening_id = sc.id
		 WHERE sc.job_id = ?`, jobID).Scan(&count)
	return count, err
}

// User operations

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	var twoFactorSecret sql.NullString
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, role, two_factor_secret, active, last_login_at, created_at, updated_at
		 FROM users WHERE email = ?`, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &twoFactorSecret, &u.Active, &lastLoginAt, &u.CreatedAt, &u.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		fmt.Printf("DB Error in GetUserByEmail: %v\n", err)
		return nil, err
	}

	if twoFactorSecret.Valid {
		u.TwoFactorSecret = twoFactorSecret.String
	}
	if lastLoginAt.Valid {
		t := lastLoginAt.Time
		u.LastLoginAt = &t
	}

	return &u, nil
}

func (r *Repository) UpdateSanctionListCount(ctx context.Context, listID int64, count int) error {
	query := `UPDATE sanction_lists SET record_count = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, count, listID)
	return err
}

func (r *Repository) UpdateUserLastLogin(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET last_login_at = CURRENT_TIMESTAMP WHERE id = ?`, userID)
	return err
}

// ResolveMatches maps hashes back to customer and sanction records
func (r *Repository) ResolveMatches(ctx context.Context, hashes []uint64, customerListID int64, sanctionListIDs []int64) ([]struct {
	Customer models.Customer
	Sanction models.Sanction
}, error) {
	// Create hash map for efficient lookup
	hashMap := make(map[uint64]bool)
	for _, h := range hashes {
		hashMap[h] = true
	}

	// Get matching customers
	customers, err := r.GetCustomersByListID(ctx, customerListID)
	if err != nil {
		return nil, err
	}
	customerMap := make(map[uint64]models.Customer)
	for _, c := range customers {
		if hashMap[c.Hash] {
			customerMap[c.Hash] = c
		}
	}

	// Get matching sanctions
	sanctions, err := r.GetSanctionsByListIDs(ctx, sanctionListIDs)
	if err != nil {
		return nil, err
	}
	sanctionMap := make(map[uint64]models.Sanction)
	for _, s := range sanctions {
		if hashMap[s.Hash] {
			sanctionMap[s.Hash] = s
		}
	}

	// Pair them up
	var matches []struct {
		Customer models.Customer
		Sanction models.Sanction
	}
	for hash := range hashMap {
		if c, cOk := customerMap[hash]; cOk {
			if s, sOk := sanctionMap[hash]; sOk {
				matches = append(matches, struct {
					Customer models.Customer
					Sanction models.Sanction
				}{Customer: c, Sanction: s})
			}
		}
	}

	return matches, nil
}

// GetDashboardStats returns aggregated statistics for the dashboard
func (r *Repository) GetDashboardStats(ctx context.Context) (int64, int64, int64, []*models.Screening, error) {
	var totalScreenings int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM screenings").Scan(&totalScreenings); err != nil {
		return 0, 0, 0, nil, err
	}

	var totalMatches int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM screening_results").Scan(&totalMatches); err != nil {
		return 0, 0, 0, nil, err
	}

	var activeLists int64
	var customerLists int64
	var sanctionLists int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM customer_lists").Scan(&customerLists); err != nil {
		return 0, 0, 0, nil, err
	}
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sanction_lists").Scan(&sanctionLists); err != nil {
		return 0, 0, 0, nil, err
	}
	activeLists = customerLists + sanctionLists

	// Get recent screenings (top 5)
	rows, err := r.db.QueryContext(ctx, 
		`SELECT id, job_id, name, status, match_count, finished_at, created_at 
		 FROM screenings ORDER BY created_at DESC LIMIT 5`)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	defer rows.Close()

	var recentScreenings []*models.Screening
	for rows.Next() {
		var s models.Screening
		var finishedAt sql.NullTime
		if err := rows.Scan(&s.ID, &s.JobID, &s.Name, &s.Status, &s.MatchCount, &finishedAt, &s.CreatedAt); err != nil {
			return 0, 0, 0, nil, err
		}
		if finishedAt.Valid {
			s.FinishedAt = finishedAt.Time
		}
		recentScreenings = append(recentScreenings, &s)
	}

	return totalScreenings, totalMatches, activeLists, recentScreenings, nil
}
