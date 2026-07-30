package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// --- POST /register ---

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// --- POST /login ---

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// --- GET /health ---
type HealthResponse struct {
	Status string `json:"status"`
}

// --- POST /verify-email ---
type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type VerifyEmailResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// --- POST /resend-verification ---
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

type ResendVerificationReponse struct {
	Message string `json:"message"`
}

// -- shared ---
type ErrorResponse struct {
	Error string `json:"error"`
}

// helper handler
type BaseHandler struct{}

func (b *BaseHandler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("handlers: encode JSON response: %v", err)
	}
}

func (b *BaseHandler) writeError(w http.ResponseWriter, status int, msg string) {
	b.writeJSON(w, status, ErrorResponse{Error: msg})
}

type RegisterHandler struct {
	BaseHandler
}

// --- Handlers ---
func (rh *RegisterHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rh.writeError(w, http.StatusBadRequest, "Unable to decode Json")
	}
	defer r.Body.Close()

	log.Println("Registering User: ", req.Username, req.Email)

	rh.writeJSON(w, http.StatusCreated, RegisterResponse{
		ID:     "dml-uuid-123",
		Status: "pending_verification",
	})

}

func (rh *RegisterHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rh.writeError(w, http.StatusBadRequest, "Unable to decode Json")
	}
	defer r.Body.Close()

	log.Println("Loging in with User: ", req.Username)

	rh.writeJSON(w, http.StatusOK, LoginResponse{
		ID:     "dml-uuid-456",
		Status: "active",
	})
}

func (rh *RegisterHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	rh.writeJSON(w, http.StatusOK, HealthResponse{
		Status: "healthy 100%",
	})
}

func (rh *RegisterHandler) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {

}

func (rh *RegisterHandler) HandleResendVerification(w http.ResponseWriter, r *http.Request) {

}

func serverHandling() {

	rh := &RegisterHandler{}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", rh.HandleRegister)
	mux.HandleFunc("POST /login", rh.HandleLogin)
	mux.HandleFunc("GET /health", rh.HandleHealth)
	mux.HandleFunc("POST /verify-email", rh.HandleVerifyEmail)
	mux.HandleFunc("POST /resend-verification", rh.HandleResendVerification)

	// Start server
	addr := ":8080"
	log.Printf("Server starting on %s", addr)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}

func main() {
	serverHandling()
}
