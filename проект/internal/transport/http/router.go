package http

import (
    "net/http"
    
    "github.com/gorilla/mux"
    "github.com/opr1234/calculator/internal/auth"
)

func NewRouter(h *Handler, authMiddleware mux.MiddlewareFunc) *mux.Router {
    r := mux.NewRouter()
    
    r.Use(mux.CORSMethodMiddleware(r))
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            next.ServeHTTP(w, r)
        }
    })

    public := r.PathPrefix("/api/v1").Subrouter()
    public.HandleFunc("/register", h.Register).Methods("POST", "OPTIONS")
    public.HandleFunc("/login", h.Login).Methods("POST", "OPTIONS")

    protected := r.PathPrefix("/api/v1").Subrouter()
    protected.Use(authMiddleware)
    protected.HandleFunc("/calculate", h.Calculate).Methods("POST", "OPTIONS")

    r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        sendError(w, http.StatusNotFound, "Endpoint not found")
    })

    return r
}
