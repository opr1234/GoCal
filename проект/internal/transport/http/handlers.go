package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/yourusername/calculator/internal/auth"
	"github.com/yourusername/calculator/internal/storage"
	pb "github.com/yourusername/calculator/proto"
)

type Handler struct {
	storage    *storage.Storage
	calculator pb.CalculatorClient
	secret     string
}

func NewHandler(
	storage *storage.Storage,
	calculator pb.CalculatorClient,
	secret string,
) *Handler {
	return &Handler{
		storage:    storage,
		calculator: calculator,
		secret:     secret,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Login == "" || req.Password == "" {
		sendError(w, http.StatusBadRequest, "Login and password are required")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if _, err := h.storage.CreateUser(req.Login, hash); err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			sendError(w, http.StatusConflict, "User already exists")
			return
		}
		sendError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	user, err := h.storage.GetUserByLogin(req.Login)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			sendError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		sendError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		sendError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := auth.GenerateToken(user.ID, h.secret)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func (h *Handler) Calculate(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.ContextKeyUserID).(int)

	var req struct {
		Expression string `json:"expression"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if !isValidExpression(req.Expression) {
		sendError(w, http.StatusUnprocessableEntity, "Invalid expression")
		return
	}

	exprID, err := h.storage.SaveExpression(userID, req.Expression)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	go h.processExpression(r.Context(), exprID, userID, req.Expression)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     exprID,
		"status": "pending",
	})
}

func (h *Handler) processExpression(
	ctx context.Context,
	exprID int64,
	userID int,
	expression string,
) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	res, err := h.calculator.Evaluate(ctx, &pb.ExpressionRequest{
		Expression: expression,
		UserId:     int32(userID),
	})

	var status string
	var result float64
	if err != nil {
		status = "error"
	} else if res.Error != "" {
		status = "error"
	} else {
		status = "completed"
		result = res.Result
	}

	if err := h.storage.UpdateExpressionStatus(exprID, status, result); err != nil {
		log.Printf("Failed to update expression status: %v", err)
	}
}

func isValidExpression(expr string) bool {
	return true
}

func sendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
