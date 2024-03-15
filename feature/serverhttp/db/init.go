package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func CreateConnection(databaseFile string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func CreateDB(conn *sql.DB) error {
	createContentTable := `
		CREATE TABLE IF NOT EXISTS content (
			id INTEGER PRIMARY KEY,
			location TEXT NOT NULL,
			content BLOB,
			cookie TEXT
		);
	`
	_, err := conn.Exec(createContentTable)
	if err != nil {
		return err
	}
	return nil
}

func InsertContent(conn *sql.DB, location, content string, cookie string) (int64, error) {
	sql := `
		INSERT INTO content(location, content, cookie)
		VALUES(?, ?, ?)
	`
	result, err := conn.Exec(sql, location, content, cookie)
	if err != nil {
		return 0, err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastID, nil
}

func GetContent(conn *sql.DB, location string) (string, string, error) {
	sql := `
		SELECT content
		FROM content
		WHERE location = ?
	`
	var content string
	err := conn.QueryRow(sql, location).Scan(&content)
	if err != nil {
		return "", "", err
	}
	sql2 := `
		SELECT cookie
		FROM content
		WHERE location = ?
	`
	var cookie string
	err = conn.QueryRow(sql2, location).Scan(&cookie)
	if err != nil {
		return "", "", err
	}
	return content, cookie, nil
}

func GetLocations(conn *sql.DB) ([]string, error) {
	sql := `
		SELECT DISTINCT location
		FROM content
	`
	rows, err := conn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []string
	for rows.Next() {
		var location string
		err := rows.Scan(&location)
		if err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}
	return locations, nil
}
