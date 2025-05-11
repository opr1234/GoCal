package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrExpressionNotFound = errors.New("expression not found")
)

type Storage struct {
	db *sql.DB
}

type User struct {
	ID           int
	Login        string
	PasswordHash string
}

type Expression struct {
	ID         int64
	UserID     int
	Expression string
	Status     string
	Result     float64
	CreatedAt  time.Time
}

func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        login TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL
    );
    
    CREATE TABLE IF NOT EXISTS expressions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        expression TEXT NOT NULL,
        status TEXT NOT NULL DEFAULT 'pending',
        result REAL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(user_id) REFERENCES users(id)
    );
    
    CREATE INDEX IF NOT EXISTS idx_expressions_user ON expressions(user_id);
    
    `
	_, err := s.db.Exec(query)
	return err
}

func (s *Storage) CreateUser(login, passwordHash string) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("transaction begin failed: %w", err)
	}
	defer tx.Rollback()

	var existingID int
	err = tx.QueryRow("SELECT id FROM users WHERE login = ?", login).Scan(&existingID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("user check failed: %w", err)
	}
	if existingID > 0 {
		return 0, ErrUserExists
	}

	res, err := tx.Exec(
		"INSERT INTO users (login, password_hash) VALUES (?, ?)",
		login, passwordHash,
	)
	if err != nil {
		return 0, fmt.Errorf("user insert failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("transaction commit failed: %w", err)
	}

	return res.LastInsertId()
}

func (s *Storage) GetUserByLogin(login string) (*User, error) {
	var user User
	err := s.db.QueryRow(
		"SELECT id, login, password_hash FROM users WHERE login = ?",
		login,
	).Scan(&user.ID, &user.Login, &user.PasswordHash)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user query failed: %w", err)
	}

	return &user, nil
}

func (s *Storage) SaveExpression(userID int, expr string) (int64, error) {
	res, err := s.db.Exec(
		"INSERT INTO expressions (user_id, expression) VALUES (?, ?)",
		userID, expr,
	)
	if err != nil {
		return 0, fmt.Errorf("expression insert failed: %w", err)
	}

	return res.LastInsertId()
}

func (s *Storage) UpdateExpressionStatus(id int64, status string, result float64) error {
	_, err := s.db.Exec(
		"UPDATE expressions SET status = ?, result = ? WHERE id = ?",
		status, result, id,
	)
	return err
}

func (s *Storage) GetUserExpressions(userID int) ([]Expression, error) {
	rows, err := s.db.Query(
		"SELECT id, expression, status, result, created_at FROM expressions WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("expressions query failed: %w", err)
	}
	defer rows.Close()
	var expressions []Expression
	for rows.Next() {
		var expr Expression
		if err := rows.Scan(
			&expr.ID,
			&expr.Expression,
			&expr.Status,
			&expr.Result,
			&expr.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("expression scan failed: %w", err)
		}
		expr.UserID = userID
		expressions = append(expressions, expr)
	}

	return expressions, nil
}

func (s *Storage) GetPendingExpressions() ([]Expression, error) {
	rows, err := s.db.Query(
		"SELECT id, user_id, expression FROM expressions WHERE status = 'pending'",
	)
	if err != nil {
		return nil, fmt.Errorf("pending expressions query failed: %w", err)
	}
	defer rows.Close()

	var expressions []Expression
	for rows.Next() {
		var expr Expression
		if err := rows.Scan(
			&expr.ID,
			&expr.UserID,
			&expr.Expression,
		); err != nil {
			return nil, fmt.Errorf("pending expression scan failed: %w", err)
		}
		expressions = append(expressions, expr)
	}

	return expressions, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
