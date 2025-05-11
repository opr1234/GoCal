package models

import (
	"time"
)

type User struct {
	ID           int       `json:"id" db:"id"`
	Login        string    `json:"login" db:"login"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Expression struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int       `json:"user_id" db:"user_id"`
	Expression string    `json:"expression" db:"expression"`
	Status     string    `json:"status" db:"status"`
	Result     float64   `json:"result,omitempty" db:"result"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type APIError struct {
	Error      string `json:"error"`
	StatusCode int    `json:"-"`
}

type JWTResponse struct {
	Token string `json:"token"`
}

type CalculationRequest struct {
	Expression string `json:"expression"`
}

type CalculationResponse struct {
	ID     int64   `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result,omitempty"`
}
