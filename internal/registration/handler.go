package registration

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	credentialv1 "github.com/minhloi0901/go-password-regis/internal/genproto/credential/v1"
	"github.com/minhloi0901/go-password-regis/internal/prospect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// helper handler
type BaseHandler struct {
	validate *validator.Validate
}

type RegisterHandler struct {
	BaseHandler

	ProspectRepo      prospect.Repository
	CredentialService credentialv1.CredentialServiceClient
}

func NewRegisterHandler(prospectRepo prospect.Repository, credentialService credentialv1.CredentialServiceClient) *RegisterHandler {
	return &RegisterHandler{
		BaseHandler: BaseHandler{
			validate: validator.New(),
		},
		ProspectRepo:      prospectRepo,
		CredentialService: credentialService,
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

	log.Println("Credential Request: ", credentialRequest.Username)
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
