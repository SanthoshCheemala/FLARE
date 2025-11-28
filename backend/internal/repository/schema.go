package repository

const SQLiteSchema = `
CREATE TABLE IF NOT EXISTS customer_lists (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    file_path TEXT,
    record_count INTEGER DEFAULT 0,
    uploaded_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS customers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    external_id TEXT NOT NULL,
    name TEXT NOT NULL,
    dob TEXT,
    country TEXT,
    hash INTEGER NOT NULL,
    list_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (list_id) REFERENCES customer_lists(id)
);

CREATE TABLE IF NOT EXISTS sanction_lists (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    source TEXT NOT NULL,
    description TEXT,
    file_path TEXT,
    record_count INTEGER DEFAULT 0,
    version INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sanctions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT NOT NULL,
    name TEXT NOT NULL,
    dob TEXT,
    country TEXT,
    program TEXT,
    hash INTEGER NOT NULL,
    list_id INTEGER NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    version INTEGER DEFAULT 1,
    FOREIGN KEY (list_id) REFERENCES sanction_lists(id)
);

CREATE TABLE IF NOT EXISTS screenings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    customer_list_id INTEGER NOT NULL,
    sanction_list_ids TEXT NOT NULL,
    status TEXT NOT NULL,
    match_count INTEGER DEFAULT 0,
    customer_count INTEGER DEFAULT 0,
    sanction_count INTEGER DEFAULT 0,
    worker_count INTEGER DEFAULT 0,
    memory_estimate_mb REAL DEFAULT 0,
    started_at DATETIME,
    finished_at DATETIME,
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_list_id) REFERENCES customer_lists(id)
);

CREATE TABLE IF NOT EXISTS screening_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    screening_id INTEGER NOT NULL,
    customer_id INTEGER NOT NULL,
    sanction_id INTEGER NOT NULL,
    match_score REAL NOT NULL,
    status TEXT NOT NULL,
    investigator_id INTEGER,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (screening_id) REFERENCES screenings(id),
    FOREIGN KEY (customer_id) REFERENCES customers(id),
    FOREIGN KEY (sanction_id) REFERENCES sanctions(id)
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    two_factor_secret TEXT,
    active INTEGER DEFAULT 1,
    last_login_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    actor_id INTEGER NOT NULL,
    action TEXT NOT NULL,
    entity_type TEXT,
    entity_id TEXT,
    details TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create default admin user (password: admin123)
INSERT OR IGNORE INTO users (id, email, password_hash, role, active) 
VALUES (1, 'admin@flare.local', '$2a$10$8K1p/a0dL3LKzN8kHbLNEOq9ZG5Cx3HV8/7NULFUfLNDIGNqK7Nda', 'admin', 1);

-- Create sample customer list
INSERT OR IGNORE INTO customer_lists (id, name, description, record_count, uploaded_by)
VALUES (1, 'Sample Customers', 'Test customer data', 3, 1);

-- Create sample customers
INSERT OR IGNORE INTO customers (external_id, name, dob, country, hash, list_id) VALUES
('CUST001', 'John Doe', '1980-01-15', 'US', 123456789, 1),
('CUST002', 'Jane Smith', '1975-05-20', 'UK', 987654321, 1),
('CUST003', 'Bob Johnson', '1990-03-10', 'CA', 456789123, 1);

-- Create sample sanction list
INSERT OR IGNORE INTO sanction_lists (id, name, source, description, record_count)
VALUES (1, 'OFAC SDN', 'OFAC', 'Office of Foreign Assets Control Specially Designated Nationals', 2);

-- Create sample sanctions
INSERT OR IGNORE INTO sanctions (source, name, dob, country, program, hash, list_id) VALUES
('OFAC', 'John Doe', '1980-01-15', 'US', 'SANCTIONS', 123456789, 1),
('OFAC', 'Evil Corp', '1990-01-01', 'XX', 'TERRORISM', 111222333, 1);
`

func (r *Repository) InitSchema() error {
	if _, err := r.db.Exec(SQLiteSchema); err != nil {
		return err
	}

	// Migrations: Attempt to add file_path column if it doesn't exist
	// We ignore errors here because if the column exists, it will fail, which is fine for this dev setup.
	// In a production system, we would use a proper migration tool.
	r.db.Exec(`ALTER TABLE customer_lists ADD COLUMN file_path TEXT`)
	r.db.Exec(`ALTER TABLE sanction_lists ADD COLUMN file_path TEXT`)

	return nil
}
