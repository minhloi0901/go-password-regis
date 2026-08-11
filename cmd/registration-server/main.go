package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/minhloi0901/go-password-regis/internal/config"
	credentialv1 "github.com/minhloi0901/go-password-regis/internal/genproto/credential/v1"
	"github.com/minhloi0901/go-password-regis/internal/prospect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// --- POST /register ---

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=5,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=80"`
}

type RegisterResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// --- POST /login ---

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=5,max=20"`
	// Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=80"`
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
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,min=6,max=6"`
}

type VerifyEmailResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// --- POST /resend-verification ---
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResendVerificationReponse struct {
	Message string `json:"message"`
}

// -- shared ---
type ErrorResponse struct {
	Error string `json:"error"`
}

// helper handler
type BaseHandler struct {
	validate *validator.Validate
}

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

	ProspectRepo      prospect.Repository
	CredentialService credentialv1.CredentialServiceClient
}

// --- Handlers ---
func (rh *RegisterHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rh.writeError(w, http.StatusBadRequest, "Unable to decode Json")
		return
	}
	defer r.Body.Close()
	if err := rh.validate.Struct(req); err != nil {
		// http.Error(w, "Validation failed", http.StatusBadRequest)
		rh.writeError(w, http.StatusBadRequest, "Validation failed")
		return
	}

	log.Println("Registering User: ", req.Username, req.Email)
	ctx := r.Context()
	if err := rh.validateUniqueProspect(ctx, req.Email, req.Username); err != nil {
		switch {
		case errors.Is(err, prospect.ErrEmailTaken):
			rh.writeError(w, http.StatusConflict, "email already existed")
		case errors.Is(err, prospect.ErrUsernameTaken):
			rh.writeError(w, http.StatusConflict, "username already existed")
		default:
			rh.writeError(w, http.StatusInternalServerError, "could not check prospect repository")
		}
		return
	}

	// time.Sleep(10 * time.Second)

	prospectId, err := rh.ProspectRepo.Insert(ctx, req.Username, req.Email)
	if err != nil {
		rh.writeError(w, http.StatusInternalServerError, "could not register when creating prospect")
		return
	}

	credentialRequest := &credentialv1.CreateCredentialRequest{
		ProspectId: prospectId,
		Username:   req.Username,
		Password:   req.Password,
	}
	log.Println("Credential Request: ", credentialRequest)
	if _, err = rh.CredentialService.CreateCredential(ctx, credentialRequest); err != nil {
		if st, ok := status.FromError(err); !ok {
			rh.writeError(w, http.StatusInternalServerError, "could not register when creating credentials")
		} else {
			rh.writeError(w, grpcCodeToHttpStatus(st.Code()), st.Message())
		}
		return
	}

	rh.writeJSON(w, http.StatusCreated, RegisterResponse{
		ID:     prospectId,
		Status: "pending",
	})
}

func grpcCodeToHttpStatus(code codes.Code) int {
	switch code {
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.NotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func (rh *RegisterHandler) validateUniqueProspect(ctx context.Context, email, username string) error {
	exists, err := rh.ProspectRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return prospect.ErrEmailTaken
	}

	exists, err = rh.ProspectRepo.ExistsByUsername(ctx, username)
	if err != nil {
		// rh.writeError(w, http.StatusInternalServerError, "could not check prospect repository")
		return err
	}
	if exists {
		return prospect.ErrUsernameTaken
	}
	return nil
}

// not implement
func (rh *RegisterHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rh.writeError(w, http.StatusBadRequest, "Unable to decode Json")
		return
	}
	defer r.Body.Close()
	if err := rh.validate.Struct(req); err != nil {
		rh.writeError(w, http.StatusBadRequest, "Validation failed")
		return
	}

	log.Println("Loging in with User Email: ", req.Username)

	rh.writeJSON(w, http.StatusOK, LoginResponse{
		ID:     "dml-uuid-456",
		Status: "active",
	})
}

func (rh *RegisterHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	rh.writeJSON(w, http.StatusOK, HealthResponse{
		Status: "healthy",
	})
}

func (rh *RegisterHandler) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {

}

func (rh *RegisterHandler) HandleResendVerification(w http.ResponseWriter, r *http.Request) {

}

func NewRouter(rh *RegisterHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", rh.HandleRegister)
	mux.HandleFunc("POST /login", rh.HandleLogin)
	mux.HandleFunc("GET /health", rh.HandleHealth)
	mux.HandleFunc("POST /verify-email", rh.HandleVerifyEmail)
	mux.HandleFunc("POST /resend-verification", rh.HandleResendVerification)

	return mux
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// start postgre connection
	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN()), &gorm.Config{TranslateError: true})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	// start credential grpc connection
	credentialConn, err := grpc.NewClient(
		cfg.CredentialServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // tls connection
	)
	if err != nil {
		log.Fatalf("cannot connect grpc port: %s with error %v", cfg.CredentialServiceAddr, err)
	}
	defer credentialConn.Close()

	rh := &RegisterHandler{
		BaseHandler: BaseHandler{
			validate: validator.New(),
		},
		ProspectRepo:      prospect.NewPostgresRespository(db),
		CredentialService: credentialv1.NewCredentialServiceClient(credentialConn),
	}

	// start registration http
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: NewRouter(rh),
	}
	log.Println("Registration Server is running on port: ", cfg.HTTPAddr)
	log.Fatal(server.ListenAndServe())
}
