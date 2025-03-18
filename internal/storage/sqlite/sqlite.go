package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	resp "ex.com/internal/lib/api/response"
	"ex.com/internal/storage"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Save(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveUrl"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ERR_URL_EXISTS)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last id", op)
	}

	return id, nil

}

func (s *Storage) GetUrlByAlias(alias string) (string, error) {
	const op = "storage.sqlite.GetUrlByAlias"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	var resURL string

	err = stmt.QueryRow(alias).Scan(&resURL)
	if errors.Is(err, sql.ErrNoRows) {
		return " ", fmt.Errorf("%s: %w", op, storage.ERR_URL_NOT_FOUND)
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteUrlByAlias(alias string) (resp.Url, error) {
	const op = "storage.sqlite.DeleteUrlByAlias"
	var result resp.Url
	err := s.db.QueryRow("SELECT id, alias, url FROM url WHERE alias = $1", alias).Scan(&result.Id, &result.Alias, &result.Url)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return resp.Url{}, fmt.Errorf("%s: %w", op, storage.ERR_URL_NOT_FOUND)
		}

		return resp.Url{}, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")

	if err != nil {
		return resp.Url{}, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(alias)

	if errors.Is(err, sql.ErrNoRows) {
		return resp.Url{}, storage.ERR_URL_NOT_FOUND
	}

	if err != nil {
		return resp.Url{}, fmt.Errorf("%s, %w", op, err)
	}

	return result, nil
}

func (s *Storage) GetAll() ([]resp.Url, error) {
	const op = "storage.sqlite.GetAll"
	rows, err := s.db.Query("SELECT id, url, alias FROM url")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	var result []resp.Url
	for rows.Next() {
		var id int64
		var url string
		var alias string
		if err := rows.Scan(&id, &url, &alias); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		result = append(result, resp.Url{
			Id:    id,
			Alias: alias,
			Url:   url,
		})
	}
	return result, nil

}
