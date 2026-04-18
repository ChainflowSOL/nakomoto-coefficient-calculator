package main

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type HistoryRecord struct {
	ChainToken string `json:"chain_token"`
	ChainName  string `json:"chain_name"`
	NCValue    int    `json:"nc_value"`
	Timestamp  string `json:"timestamp"`
}

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS nc_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chain_token TEXT NOT NULL,
		chain_name TEXT NOT NULL,
		nc_value INTEGER NOT NULL,
		timestamp DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_nc_history_chain_token ON nc_history(chain_token);
	CREATE INDEX IF NOT EXISTS idx_nc_history_timestamp ON nc_history(timestamp);
	`

	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}

	log.Println("NC history database initialized at", dbPath)
	return db, nil
}

func SaveNCSnapshot(db *sql.DB, coefficients []JsonResponse) error {
	if db == nil {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO nc_history (chain_token, chain_name, nc_value, timestamp) VALUES (?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)

	for _, c := range coefficients {
		_, err := stmt.Exec(c.ChainToken, c.ChainName, c.NakaCoCurrVal, now)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Printf("📊 Saved NC snapshot for %d chains", len(coefficients))
	return nil
}

func GetNCHistory(db *sql.DB, chainToken string, days int) ([]HistoryRecord, error) {
	var rows *sql.Rows
	var err error

	if days > 0 {
		cutoff := time.Now().UTC().AddDate(0, 0, -days).Format(time.RFC3339)
		rows, err = db.Query(
			"SELECT chain_token, chain_name, nc_value, timestamp FROM nc_history WHERE chain_token = ? AND timestamp >= ? ORDER BY timestamp ASC",
			chainToken, cutoff,
		)
	} else {
		rows, err = db.Query(
			"SELECT chain_token, chain_name, nc_value, timestamp FROM nc_history WHERE chain_token = ? ORDER BY timestamp ASC",
			chainToken,
		)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []HistoryRecord
	for rows.Next() {
		var r HistoryRecord
		if err := rows.Scan(&r.ChainToken, &r.ChainName, &r.NCValue, &r.Timestamp); err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	return records, nil
}

func GetAllChainsLatestHistory(db *sql.DB, days int) ([]HistoryRecord, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -days).Format(time.RFC3339)

	rows, err := db.Query(
		"SELECT chain_token, chain_name, nc_value, timestamp FROM nc_history WHERE timestamp >= ? ORDER BY chain_token ASC, timestamp ASC",
		cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []HistoryRecord
	for rows.Next() {
		var r HistoryRecord
		if err := rows.Scan(&r.ChainToken, &r.ChainName, &r.NCValue, &r.Timestamp); err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	return records, nil
}