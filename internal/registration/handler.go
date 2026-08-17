package registration

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/minhloi0901/go-password-regis/internal/email"
	credentialv1 "github.com/minhloi0901/go-password-regis/internal/genproto/credential/v1"
	"github.com/minhloi0901/go-password-regis/internal/prospect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const credentialCallTimeout = 5 * time.Second

// helper handler
type BaseHandler struct {
	validate *validator.Validate
}

type RegisterHandler struct {
	BaseHandler

	ProspectRepo      prospect.Repository
	CredentialService credentialv1.CredentialServiceClient
	EmailService      email.EmailService
}

func NewRegisterHandler(prospectRepo prospect.Repository, credentialService credentialv1.CredentialServiceClient, emailService email.EmailService) *RegisterHandler {
	return &RegisterHandler{
		BaseHandler: BaseHandler{
			validate: validator.New(),
		},
		ProspectRepo:      prospectRepo,
		CredentialService: credentialService,
		EmailService:      emailService,
	}
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
			log.Printf("Register - Pre-check Unique: %v", err)
			rh.writeError(w, http.StatusInternalServerError, "could not check prospect repository")
		}
		return
	}

	verificationCode, err := prospect.GenerateVerificationCode()
	if err != nil {
		log.Printf("Register - Generate Verification Code: %v", err)
		rh.writeError(w, http.StatusInternalServerError, "could not register when generating verification code")
		return
	}
	codeExpiresAt := time.Now().Add(prospect.VerificationCodeTTL)
	expiresAt := time.Now().Add(prospect.ProspectTTL)

	// save prospect
	prospectId, err := rh.ProspectRepo.Insert(ctx, req.Username, req.Email, verificationCode, codeExpiresAt, expiresAt)
	if err != nil {
		if errors.Is(err, prospect.ErrProspectConflict) {
			rh.writeError(w, http.StatusConflict, "email or username already exists")
			return
		}
		log.Printf("Register - Insert Prospect: %v", err)
		rh.writeError(w, http.StatusInternalServerError, "could not register when creating prospect")
		return
	}

	credentialRequest := &credentialv1.CreateCredentialRequest{
		ProspectId: prospectId,
		Username:   req.Username,
		Password:   req.Password,
	}

	log.Println("Credential Request: ", credentialRequest.Username)

	ctx, cancel := context.WithTimeout(ctx, credentialCallTimeout)
	defer cancel()

	// save credential
	if _, err = rh.CredentialService.CreateCredential(ctx, credentialRequest); err != nil {
		// undo the insert prospect operation if CreateCredential failed
		rh.deleteProspect(ctx, prospectId)

		if st, ok := status.FromError(err); !ok {
			rh.writeError(w, http.StatusInternalServerError, "could not register when creating credential")
		} else {
			rh.writeError(w, grpcCodeToHttpStatus(st.Code()), st.Message())
		}
		return
	}

	// send verification code to email
	if err := rh.EmailService.SendVerificationEmail(ctx, req.Email, verificationCode); err != nil {
		log.Printf("Register - Send Verification Code to Email: %v", err)
	}

	rh.writeJSON(w, http.StatusCreated, RegisterResponse{
		ID:     prospectId,
		Status: prospect.StatusPending,
	})
}

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

	// Authentication
	ctx := r.Context()

	ctx, cancel := context.WithTimeout(ctx, credentialCallTimeout)
	defer cancel()

	credentialRequest := &credentialv1.VerifyCredentialRequest{
		Username: req.Username,
		Password: req.Password,
	}

	log.Println("Verify Credential Request: ", credentialRequest.Username)

	credentialResponse, err := rh.CredentialService.VerifyCredential(ctx, credentialRequest)
	if err != nil {
		if st, ok := status.FromError(err); !ok {
			rh.writeError(w, http.StatusInternalServerError, "could not register when verifying credential")
		} else {
			rh.writeError(w, grpcCodeToHttpStatus(st.Code()), st.Message())
		}
		return
	}

	if !credentialResponse.Valid {
		rh.writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	log.Println("Log in with User Email: ", req.Username)

	// Authorization
	currentProspect, err := rh.ProspectRepo.FindById(ctx, credentialResponse.ProspectId)
	if err != nil {
		log.Printf("Login - Find Prospect: %v", err)
		rh.writeError(w, http.StatusInternalServerError, "could not login")
		return
	}
	if currentProspect.Status != prospect.StatusActive {
		switch currentProspect.Status {
		case prospect.StatusPending:
			rh.writeError(w, http.StatusForbidden, "account not verified")
		case prospect.StatusSuspended:
			rh.writeError(w, http.StatusForbidden, "account suspended")
		default:
			rh.writeError(w, http.StatusForbidden, "account not active yet")
		}
		return
	}

	log.Println("Login successful for user: ", currentProspect.Username)

	rh.writeJSON(w, http.StatusOK, LoginResponse{
		ID:     currentProspect.ID,
		Status: currentProspect.Status,
	})
}

func (rh *RegisterHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	rh.writeJSON(w, http.StatusOK, HealthResponse{
		Status: "healthy",
	})
}

func (rh *RegisterHandler) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rh.writeError(w, http.StatusBadRequest, "Unable to decode Json")
		return
	}
	defer r.Body.Close()
	if err := rh.validate.Struct(req); err != nil {
		rh.writeError(w, http.StatusBadRequest, "Validation failed")
		return
	}

	ctx := r.Context()
	// find prospect
	tempProspect, err := rh.ProspectRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		log.Printf("Verify Email - Find Prospect: %v", err)
		rh.writeError(w, http.StatusInternalServerError, "could not verify")
		return
	}

	// check status of prospect
	if tempProspect.Status == prospect.StatusActive {
		rh.writeJSON(w, http.StatusOK, VerifyEmailResponse{
			ID:     tempProspect.ID,
			Status: tempProspect.Status,
		})
		return
	}
	if tempProspect.Status != prospect.StatusPending {
		rh.writeError(w, http.StatusForbidden, "account cannot be verified")
		return
	}

	// verify code
	if err := prospect.ValidateVerificationCode(tempProspect, req.Code); err != nil {
		log.Printf("Verfiy Email - Validate Code: %v", err)
		rh.writeError(w, http.StatusBadRequest, "invalid or expired verification code")
		return
	}

	// active prospect if verify successful
	if err := rh.ProspectRepo.Active(ctx, tempProspect.ID); err != nil {
		log.Printf("Verify Email - Active Prospect: %v", err)
		rh.writeError(w, http.StatusInternalServerError, "could not veify")
	}

	log.Println("Email verified succesfully for user: ", tempProspect.Username)

	rh.writeJSON(w, http.StatusOK, VerifyEmailResponse{
		ID:     tempProspect.ID,
		Status: prospect.StatusActive,
	})
}

// not implement
func (rh *RegisterHandler) HandleResendVerification(w http.ResponseWriter, r *http.Request) {

}

func (rh *RegisterHandler) deleteProspect(ctx context.Context, id string) error {
	if err := rh.ProspectRepo.DeleteById(ctx, id); err != nil {
		return err
	}
	return nil
}

func grpcCodeToHttpStatus(code codes.Code) int {
	switch code {
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Unavailable:
		return http.StatusServiceUnavailable
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
