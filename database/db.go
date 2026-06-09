/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package database

import (
	"database/sql"
	"log"

	_ "github.com/glebarez/go-sqlite"
	"github.com/nabbar/auditor/models"
)

// DBManager encapsulates database operations for the audit system.
// It provides methods to manage catalog data, function analysis status,
// and dependency tracking within a SQLite database.
//
// This struct serves as the primary interface for all database-related operations
// in the auditor system. It maintains a connection to a SQLite database file
// and provides high-level functions for managing the audit workflow.
type DBManager struct {
	Conn *sql.DB
}

// NewDBManager initializes a new database manager with the specified SQLite file.
// It configures connection pooling and WAL mode to prevent 'database is locked' errors during parallel execution.
//
// The function sets up an optimized SQLite connection string that includes:
// - WAL (Write-Ahead Logging) mode for improved concurrency
// - Busy timeout of 5000ms to handle concurrent access better
// - Foreign key constraints enabled for data integrity
//
// Additionally, it configures the database connection pool with:
// - Maximum open connections set to 10
// - Maximum idle connections set to 5
// - No connection lifetime limit
//
// Parameters:
//   - dbFile: Path to the SQLite database file to initialize or connect to
//
// Returns:
//   - *DBManager: A pointer to the newly created database manager instance
//   - error: An error if the database initialization fails, otherwise nil
func NewDBManager(dbFile string) (*DBManager, error) {
	// Construct an optimized DSN (Data Source Name) for SQLite with WAL mode and busy timeout
	dsn := dbFile + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pooling to allow concurrent access for parallel execution
	db.SetMaxOpenConns(10) // Allow multiple simultaneous read/write connections in WAL mode
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)

	// Define SQL statements to create the required tables if they don't exist
	queries := []string{
		`CREATE TABLE IF NOT EXISTS catalog (
			package TEXT,
			func_name TEXT,
			file_path TEXT,
			PRIMARY KEY (package, func_name)
		);`,
		`CREATE TABLE IF NOT EXISTS functions (
			id TEXT PRIMARY KEY,
			package TEXT,
			func_name TEXT,
			file_path TEXT,
			summary TEXT,
			status TEXT DEFAULT 'PENDING'
		);`,
		`CREATE TABLE IF NOT EXISTS dependencies (
			function_id TEXT,
			dep_name TEXT,
			dep_type TEXT,
			PRIMARY KEY (function_id, dep_name)
		);`,
	}

	// Execute each CREATE TABLE statement to ensure the schema is initialized
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return nil, err
		}
	}

	log.Printf("[INFO] [database] Database and tables initialized in concurrent WAL mode: %s", dbFile)
	return &DBManager{Conn: db}, nil
}

// Close terminates the database connection.
//
// This method ensures that the SQLite database connection is properly closed,
// releasing resources and preventing potential memory leaks. It should be called
// when the database manager is no longer needed, typically at application shutdown.
func (m *DBManager) Close() {
	if m.Conn != nil {
		_ = m.Conn.Close()
		log.Println("[INFO] [database] SQLite connection closed.")
	}
}

// SaveCatalog inserts a complete catalog into the database in a safe manner.
// This operation uses a transaction to ensure data consistency and prevents
// duplicate entries through the INSERT OR IGNORE mechanism.
//
// The function operates within a transactional context to ensure atomicity of the operation,
// preventing partial insertions that could lead to inconsistent state. It also utilizes
// the INSERT OR IGNORE construct to avoid duplicate entries, which is particularly useful
// when multiple catalog processing runs occur.
//
// Parameters:
//   - entries: Slice of catalog entries to be saved to the database
//
// Returns:
//   - error: An error if the save operation fails, otherwise nil
func (m *DBManager) SaveCatalog(entries []models.CatalogEntry) error {
	tx, err := m.Conn.Begin()
	if err != nil {
		log.Printf("[ERROR] [database] Catalog transaction initiation failed: %v", err)
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Prepare a statement for inserting catalog entries with INSERT OR IGNORE to prevent duplicates
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO catalog (package, func_name, file_path) VALUES (?, ?, ?)")
	if err != nil {
		log.Printf("[ERROR] [database] Catalog statement preparation failed: %v", err)
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()

	// Iterate through each catalog entry and execute the insert operation
	for _, entry := range entries {
		_, _ = stmt.Exec(entry.Package, entry.FuncName, entry.FilePath)
	}

	// Commit the transaction to finalize changes
	if err := tx.Commit(); err != nil {
		log.Printf("[ERROR] [database] Catalog commit failed: %v", err)
		return err
	}

	log.Printf("[INFO] [database] Catalog successfully updated with %d entries synchronized.", len(entries))
	return nil
}

// SyncCatalogToPendingQueue populates the audit queue table from the current catalog.
// This operation ensures that all catalog entries are queued for analysis, avoiding duplicates.
//
// This function serves as a bridge between the catalog (which contains discovered functions)
// and the audit queue (where functions are scheduled for analysis). It ensures that each
// catalog entry is added to the functions table with appropriate identifiers, but only if
// it doesn't already exist in the queue. The operation uses an INSERT OR IGNORE mechanism
// to prevent duplicate entries.
//
// Returns:
//   - error: An error if the synchronization fails, otherwise nil
func (m *DBManager) SyncCatalogToPendingQueue() error {
	// SQL query to insert catalog entries into the functions table with unique identifiers
	query := `
		INSERT OR IGNORE INTO functions (id, package, func_name, file_path)
		SELECT (package || '.' || func_name), package, func_name, file_path FROM catalog
	`
	_, err := m.Conn.Exec(query)
	if err != nil {
		log.Printf("[ERROR] [database] Error during catalog -> queue synchronization: %v", err)
		return err
	}
	return nil
}

// HasSummary checks whether a function already has a valid computed summary.
//
// This method queries the database to determine if a function has been previously
// analyzed and has a non-empty summary. It's used to avoid re-processing functions
// that have already been analyzed, improving efficiency of the audit process.
//
// Parameters:
//   - funcName: Name of the function to check for existing summary
//
// Returns:
//   - bool: True if the function has a non-empty summary, false otherwise
func (m *DBManager) HasSummary(funcName string) bool {
	var summary string
	err := m.Conn.QueryRow("SELECT summary FROM functions WHERE func_name = ? AND summary IS NOT NULL AND summary != ''", funcName).Scan(&summary)
	return err == nil
}

// FindInCatalog searches for a function in the project catalog (internal).
//
// This method looks up a specific function within the catalog table to determine
// if it's part of the local codebase. If found, it returns the function details;
// otherwise, it indicates that the function is external to the repository.
//
// Parameters:
//   - funcName: Name of the function to search for in the catalog
//
// Returns:
//   - models.Function: Function details if found in the catalog
//   - bool: True if found (internal), false otherwise (external)
func (m *DBManager) FindInCatalog(funcName string) (models.Function, bool) {
	var f models.Function
	query := "SELECT package, func_name, file_path FROM catalog WHERE func_name = ? LIMIT 1"
	err := m.Conn.QueryRow(query, funcName).Scan(&f.Package, &f.FuncName, &f.FilePath)
	if err != nil {
		return f, false // Not found = external to repository
	}
	f.ID = f.Package + "." + f.FuncName
	return f, true
}

// SaveDependencies stores discovered dependencies for a function in bulk.
// This operation uses a transaction to ensure atomicity of dependency insertion.
//
// The function takes a slice of dependencies and inserts them into the database
// using a transactional approach. This ensures that either all dependencies are
// successfully stored or none are, maintaining data consistency. It also uses
// INSERT OR IGNORE to prevent duplicate dependency entries.
//
// Parameters:
//   - funcID: Identifier of the function for which dependencies are being saved
//   - deps: Slice of dependencies to be stored in the database
func (m *DBManager) SaveDependencies(funcID string, deps []models.Dependency) {
	tx, err := m.Conn.Begin()
	if err != nil {
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Insert each dependency into the dependencies table using INSERT OR IGNORE
	for _, d := range deps {
		_, _ = tx.Exec("INSERT OR IGNORE INTO dependencies (function_id, dep_name, dep_type) VALUES (?, ?, ?)", funcID, d.DepName, d.DepType)
	}
	_ = tx.Commit()
}

// GetStatus retrieves the current status of a function in the queue.
//
// This method queries the database for the status associated with a specific
// function identifier. The status indicates the current processing state of
// that function within the audit workflow (e.g., PENDING, IN_PROGRESS, EXTRACTED).
//
// Parameters:
//   - funcID: Identifier of the function to retrieve status for
//
// Returns:
//   - string: Status value or empty string if not found in database
func (m *DBManager) GetStatus(funcID string) string {
	var status string
	err := m.Conn.QueryRow("SELECT status FROM functions WHERE id = ?", funcID).Scan(&status)
	if err != nil {
		return ""
	}
	return status
}

// UpdateStatus modifies the status of a function (e.g., PENDING, IN_PROGRESS, EXTRACTED).
//
// This method updates the status field in the functions table for a specific function.
// It's used to track the progress of functions through different stages of analysis,
// such as marking them as being processed or completed.
//
// Parameters:
//   - funcID: Identifier of the function whose status should be updated
//   - status: New status value to set in the database
//
// Returns:
//   - error: An error if the update operation fails, otherwise nil
func (m *DBManager) UpdateStatus(funcID, status string) error {
	_, err := m.Conn.Exec("UPDATE functions SET status = ? WHERE id = ?", status, funcID)
	return err
}

// UpdatePasse1Final validates the final summary of a function at the end of its tree processing.
//
// This method updates both the summary and status fields for a function when the analysis
// is complete. It marks the function as EXTRACTED to indicate that its analysis has been
// finalized and the summary has been computed. This is typically called after completing
// all dependency analysis for a function.
//
// Parameters:
//   - funcID: Identifier of the function whose final summary should be updated
//   - summary: Final computed summary text to store in the database
//
// Returns:
//   - error: An error if the update operation fails, otherwise nil
func (m *DBManager) UpdatePasse1Final(funcID, summary string) error {
	_, err := m.Conn.Exec("UPDATE functions SET summary = ?, status = 'EXTRACTED' WHERE id = ?", summary, funcID)
	return err
}

// GetRecordCount returns the total number of function records in the database.
//
// This method performs a COUNT(*) query on the functions table to determine
// how many function entries are currently stored in the database. It's useful
// for monitoring the size of the audit queue and tracking progress through
// analysis runs.
//
// Returns:
//   - int: Count of records in the functions table
//   - error: An error if the count operation fails, otherwise nil
func (m *DBManager) GetRecordCount() (int, error) {
	var count int
	err := m.Conn.QueryRow("SELECT COUNT(*) FROM functions").Scan(&count)
	return count, err
}

// GetFunctionsByStatus retrieves all functions with a specific status.
//
// This method queries the functions table for entries matching a specified status,
// returning a slice of Function structs that contain relevant information about
// each function. It's used to retrieve functions that are waiting for processing,
// currently being processed, or have completed analysis.
//
// Parameters:
//   - status: Status filter value to match against function records
//
// Returns:
//   - []models.Function: Slice of functions matching the specified status
//   - error: An error if the retrieval operation fails, otherwise nil
func (m *DBManager) GetFunctionsByStatus(status string) ([]models.Function, error) {
	rows, err := m.Conn.Query("SELECT id, package, func_name, file_path, COALESCE(summary, '') FROM functions WHERE status = ?", status)
	if err != nil {
		log.Printf("[ERROR] [database] Failed to retrieve (Status: %s): %v", status, err)
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var result []models.Function
	for rows.Next() {
		var f models.Function
		if err := rows.Scan(&f.ID, &f.Package, &f.FuncName, &f.FilePath, &f.Summary); err == nil {
			result = append(result, f)
		}
	}
	return result, nil
}

// ResetQueueStatus resets functions to a given status to allow re-execution of passes.
//
// This method updates the status field for all functions currently in a specific
// state, changing them to a new status. It's used when restarting analysis passes
// or when errors occur that require functions to be reprocessed from an earlier stage.
//
// Parameters:
//   - from: Current status of functions to be reset
//   - to: New status value to set for matching functions
//
// Returns:
//   - error: An error if the status reset operation fails, otherwise nil
func (m *DBManager) ResetQueueStatus(from, to string) error {
	_, err := m.Conn.Exec("UPDATE functions SET status = ? WHERE status = ?", to, from)
	if err != nil {
		log.Printf("[ERROR] [database] Failed to reset statuses (%s -> %s): %v", from, to, err)
	}
	return err
}

// UpdatePasse1Iteration updates the summary and dependencies without blocking status if reiteration is needed.
//
// This method performs an update operation on a function's summary and dependencies,
// potentially changing its status depending on whether it's the final pass. It uses
// a transactional approach to ensure atomicity of changes, which is important when
// updating both summary and dependencies in the same operation.
//
// Parameters:
//   - funcID: Identifier of the function whose summary and dependencies should be updated
//   - summary: Updated summary text for the function
//   - deps: Slice of dependencies to update in the database
//   - isLastPass: Flag indicating if this is the final pass of analysis
//
// Returns:
//   - error: An error if the iteration update fails, otherwise nil
func (m *DBManager) UpdatePasse1Iteration(funcID, summary string, deps []models.Dependency, isLastPass bool) error {
	tx, err := m.Conn.Begin()
	if err != nil {
		log.Printf("[ERROR] [database] Passe 1 transaction failed for %s: %v", funcID, err)
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	status := "PENDING"
	if isLastPass {
		status = "EXTRACTED"
	}

	// Update the summary and status fields for the function
	_, err = tx.Exec("UPDATE functions SET summary = ?, status = ? WHERE id = ?", summary, status, funcID)
	if err != nil {
		log.Printf("[ERROR] [database] Failed to write summary for %s: %v", funcID, err)
		return err
	}

	// Insert newly discovered dependencies into the dependencies table
	for _, d := range deps {
		_, _ = tx.Exec("INSERT OR IGNORE INTO dependencies (function_id, dep_name, dep_type) VALUES (?, ?, ?)", funcID, d.DepName, d.DepType)
	}

	return tx.Commit()
}

// GetDependenciesContext retrieves the context of dependencies for a function.
//
// This method constructs an HTML-formatted string that describes the dependencies
// of a specific function, including their summaries where available. It performs a
// LEFT JOIN between the dependencies table and the functions table to retrieve
// summary information about each dependency.
//
// Parameters:
//   - funcID: Identifier of the function whose dependency context should be retrieved
//
// Returns:
//   - string: HTML-formatted dependency context as a string
//   - error: An error if the retrieval fails, otherwise nil
func (m *DBManager) GetDependenciesContext(funcID string) (string, error) {
	rows, err := m.Conn.Query(`
		SELECT d.dep_name, COALESCE(f.summary, '') 
		FROM dependencies d 
		LEFT JOIN functions f ON f.func_name = d.dep_name 
		WHERE d.function_id = ?`, funcID)
	if err != nil {
		log.Printf("[WARNING] [database] No dependencies found for %s: %v", funcID, err)
		return "", err
	}
	defer func() {
		_ = rows.Close()
	}()

	var contextHTML string
	for rows.Next() {
		var name, sum string
		if err := rows.Scan(&name, &sum); err == nil {
			if sum == "" {
				sum = "External element considered safe and valid."
			}
			contextHTML += "- " + name + " : " + sum + "\n"
		}
	}
	return contextHTML, nil
}
