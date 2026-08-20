package registration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	emailmocks "github.com/minhloi0901/go-password-regis/internal/email/mocks"
	credentialv1 "github.com/minhloi0901/go-password-regis/internal/genproto/credential/v1"
	credentialmocks "github.com/minhloi0901/go-password-regis/internal/genproto/credential/v1/mocks"
	"github.com/minhloi0901/go-password-regis/internal/prospect"
	prospectmocks "github.com/minhloi0901/go-password-regis/internal/prospect/mocks"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	testProspectId       = "prospect-id-1"
	testRegistrationBody = `{
		"username": "rudeus",
		"email": "loi.duong@example.com",
		"password": "qweasdZXC@123"
	}`
	testLoginBody = `{
		"username": "rudeus",
		"password": "qweasdZXC@123"
	}`
	testCredentialId    = "credential-id-1"
	testVerifyEmailBody = `{
		"email": "loine@example.com",
		"code": "123456"
	}`
)

var (
	testActiveProspect = prospect.Prospect{
		ID:       testProspectId,
		Username: "rudeus",
		Email:    "loi@example.com",
		Status:   prospect.StatusActive,
	}
	testPendingProspect = prospect.Prospect{
		ID:               "prospect-id-2",
		Username:         "rudy2",
		Email:            "loine@example.com",
		Status:           prospect.StatusPending,
		VerificationCode: "123456",
		CodeExpiresAt:    time.Now().Add(5 * time.Minute),
	}
	testCreateCredentialResponse = &credentialv1.CreateCredentialResponse{CredentialId: testCredentialId}
	testVerifyCredentialResponse = &credentialv1.VerifyCredentialResponse{
		Valid:      true,
		ProspectId: testProspectId,
	}
)

type MockProspectRepository struct {
	InsertFunc                 func(ctx context.Context, username, email, verificationCode string, codeExpiresAt, expiresAt time.Time) (string, error)
	ExistsByEmailFunc          func(ctx context.Context, email string) (bool, error)
	ExistsByUsernameFunc       func(ctx context.Context, username string) (bool, error)
	DeleteByIdFunc             func(ctx context.Context, id string) error
	FindByIdFunc               func(ctx context.Context, id string) (prospect.Prospect, error)
	FindByEmailFunc            func(ctx context.Context, email string) (prospect.Prospect, error)
	ActiveFunc                 func(ctx context.Context, id string) error
	UpdateVerificationCodeFunc func(ctx context.Context, id, verificationCode string, codeExpiresAt time.Time) error
}

func (m *MockProspectRepository) Insert(ctx context.Context, username, email, verificationCode string, codeExpiresAt, expiresAt time.Time) (string, error) {
	return m.InsertFunc(ctx, username, email, verificationCode, codeExpiresAt, expiresAt)
}

func (m *MockProspectRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return m.ExistsByEmailFunc(ctx, email)
}

func (m *MockProspectRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return m.ExistsByUsernameFunc(ctx, username)
}

func (m *MockProspectRepository) DeleteById(ctx context.Context, id string) error {
	return m.DeleteByIdFunc(ctx, id)
}

func (m *MockProspectRepository) FindById(ctx context.Context, id string) (prospect.Prospect, error) {
	return m.FindByIdFunc(ctx, id)
}

func (m *MockProspectRepository) FindByEmail(ctx context.Context, email string) (prospect.Prospect, error) {
	return m.FindByEmailFunc(ctx, email)
}

func (m *MockProspectRepository) Active(ctx context.Context, id string) error {
	return m.ActiveFunc(ctx, id)
}

func (m *MockProspectRepository) UpdateVerificationCode(ctx context.Context, id, verificationCode string, codeExpiresAt time.Time) error {
	return m.UpdateVerificationCodeFunc(ctx, id, verificationCode, codeExpiresAt)
}

type MockCredentialServiceClient struct {
	CreateCredentialFunc func(ctx context.Context, in *credentialv1.CreateCredentialRequest, opts ...grpc.CallOption) (*credentialv1.CreateCredentialResponse, error)
	VerifyCredentialFunc func(ctx context.Context, in *credentialv1.VerifyCredentialRequest, opts ...grpc.CallOption) (*credentialv1.VerifyCredentialResponse, error)
}

func (m *MockCredentialServiceClient) CreateCredential(ctx context.Context, in *credentialv1.CreateCredentialRequest, opts ...grpc.CallOption) (*credentialv1.CreateCredentialResponse, error) {
	return m.CreateCredentialFunc(ctx, in)
}

func (m *MockCredentialServiceClient) VerifyCredential(ctx context.Context, in *credentialv1.VerifyCredentialRequest, opts ...grpc.CallOption) (*credentialv1.VerifyCredentialResponse, error) {
	return m.VerifyCredentialFunc(ctx, in)
}

type MockEmailService struct {
	SendVerificationEmailFunc func(ctx context.Context, email, code string) error
}

func (m *MockEmailService) SendVerificationEmail(ctx context.Context, email, code string) error {
	return m.SendVerificationEmailFunc(ctx, email, code)
}

// Fill default as helper
func fillProspectDefaults(m *MockProspectRepository) {
	if m.InsertFunc == nil {
		m.InsertFunc = func(ctx context.Context, username, email, verificationCode string, codeExpiresAt, expiresAt time.Time) (string, error) {
			return testProspectId, nil
		}
	}
	if m.ExistsByEmailFunc == nil {
		m.ExistsByEmailFunc = func(ctx context.Context, email string) (bool, error) {
			return false, nil
		}
	}
	if m.ExistsByUsernameFunc == nil {
		m.ExistsByUsernameFunc = func(ctx context.Context, username string) (bool, error) {
			return false, nil
		}
	}
	if m.DeleteByIdFunc == nil {
		m.DeleteByIdFunc = func(ctx context.Context, id string) error {
			return nil
		}
	}
	if m.FindByIdFunc == nil {
		m.FindByIdFunc = func(ctx context.Context, id string) (prospect.Prospect, error) {
			return testActiveProspect, nil
		}
	}
	if m.FindByEmailFunc == nil {
		m.FindByEmailFunc = func(ctx context.Context, email string) (prospect.Prospect, error) {
			return testActiveProspect, nil
		}
	}
	if m.ActiveFunc == nil {
		m.ActiveFunc = func(ctx context.Context, id string) error {
			return nil
		}
	}
	if m.UpdateVerificationCodeFunc == nil {
		m.UpdateVerificationCodeFunc = func(ctx context.Context, id, verificationCode string, codeExpiresAt time.Time) error {
			return nil
		}
	}
}

func fillCredentialDefaults(m *MockCredentialServiceClient) {
	if m.CreateCredentialFunc == nil {
		m.CreateCredentialFunc = func(ctx context.Context, in *credentialv1.CreateCredentialRequest, opts ...grpc.CallOption) (*credentialv1.CreateCredentialResponse, error) {
			return testCreateCredentialResponse, nil
		}
	}
	if m.VerifyCredentialFunc == nil {
		m.VerifyCredentialFunc = func(ctx context.Context, in *credentialv1.VerifyCredentialRequest, opts ...grpc.CallOption) (*credentialv1.VerifyCredentialResponse, error) {
			return testVerifyCredentialResponse, nil
		}
	}
}

func fillEmailDefaults(m *MockEmailService) {
	if m.SendVerificationEmailFunc == nil {
		m.SendVerificationEmailFunc = func(ctx context.Context, email, code string) error {
			return nil
		}
	}
}

// Registration test
func TestHandleRegister(t *testing.T) {
	tests := []struct {
		name string
		body string

		mockExistsByEmail    func(ctx context.Context, email string) (bool, error)
		mockExistsByUsername func(ctx context.Context, username string) (bool, error)
		mockInsert           func(ctx context.Context, username, email, verificationCode string, codeExpiresAt, expiresAt time.Time) (string, error)
		mockDeleteById       func(ctx context.Context, id string) error
		mockCreateCredential func(ctx context.Context, in *credentialv1.CreateCredentialRequest, opts ...grpc.CallOption) (*credentialv1.CreateCredentialResponse, error)

		wantStatus     int
		wantErrContain string
		wantDelete     bool
	}{
		{
			name:       "successful registration returns 201",
			body:       testRegistrationBody,
			wantStatus: http.StatusCreated,
			wantDelete: false,
		},
		{
			name:           "username shorter than 5 returns 400",
			body:           `{"username":"abc","email":"loi@example.com","password":"AbC123456"}`,
			wantStatus:     http.StatusBadRequest,
			wantErrContain: "Validation failed",
		},
		{
			name:           "invalid email returns 400",
			body:           `{"username":"rudeus","email":"not-an-email","password":"AbC123456"}`,
			wantStatus:     http.StatusBadRequest,
			wantErrContain: "Validation failed",
		},
		{
			name:           "password shorter than 8 returns 400",
			body:           `{"username":"rudeus","email":"loi@example.com","password":"short"}`,
			wantStatus:     http.StatusBadRequest,
			wantErrContain: "Validation failed",
		},
		{
			name: "duplicate email returns 409",
			body: testRegistrationBody,
			mockExistsByEmail: func(ctx context.Context, email string) (bool, error) {
				return true, nil
			},
			wantStatus:     http.StatusConflict,
			wantErrContain: "email already existed",
		},
		{
			name: "duplicate username returns 409",
			body: testRegistrationBody,
			mockExistsByUsername: func(ctx context.Context, username string) (bool, error) {
				return true, nil
			},
			wantStatus:     http.StatusConflict,
			wantErrContain: "username already existed",
		},
		{
			name: "database failure during pre-check returns 500",
			body: testRegistrationBody,
			mockExistsByEmail: func(ctx context.Context, email string) (bool, error) {
				return false, errors.New("connection refused")
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "credential service unavailable returns 503 and delete prospect",
			body: testRegistrationBody,
			mockCreateCredential: func(ctx context.Context, in *credentialv1.CreateCredentialRequest, opts ...grpc.CallOption) (*credentialv1.CreateCredentialResponse, error) {
				return nil, status.Error(codes.Unavailable, "connection refused")
			},
			wantStatus: http.StatusServiceUnavailable,
			wantDelete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init prospect mock
			deleteCalled := false
			mockRepo := &MockProspectRepository{
				InsertFunc:           tt.mockInsert,
				ExistsByEmailFunc:    tt.mockExistsByEmail,
				ExistsByUsernameFunc: tt.mockExistsByUsername,
				DeleteByIdFunc: func(ctx context.Context, id string) error {
					deleteCalled = true
					return nil
				},
			}

			fillProspectDefaults(mockRepo)

			// init credential service mock
			mockCredentail := &MockCredentialServiceClient{
				CreateCredentialFunc: tt.mockCreateCredential,
			}

			fillCredentialDefaults(mockCredentail)

			// init email service mock
			mockEmail := &MockEmailService{}

			fillEmailDefaults(mockEmail)

			// inject mock to service
			rh := NewRegisterHandler(
				mockRepo,
				mockCredentail,
				mockEmail,
			)

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			rh.HandleRegister(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("http status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if deleteCalled != tt.wantDelete {
				t.Errorf("delete called = %v, want %v", deleteCalled, tt.wantDelete)
			}
		})
	}
}

func TestHandleLogin(t *testing.T) {
	tests := []struct {
		name string
		body string

		mockVerifyCredential func(ctx context.Context, in *credentialv1.VerifyCredentialRequest, opts ...grpc.CallOption) (*credentialv1.VerifyCredentialResponse, error)
		mockFindById         func(ctx context.Context, id string) (prospect.Prospect, error)

		wantStatus int
	}{
		{
			name:       "successful login",
			body:       testLoginBody,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid request",
			body:       `{"username":"abc","password":"short"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "wrong password",
			body: testLoginBody,
			mockVerifyCredential: func(ctx context.Context, in *credentialv1.VerifyCredentialRequest, opts ...grpc.CallOption) (*credentialv1.VerifyCredentialResponse, error) {
				return &credentialv1.VerifyCredentialResponse{Valid: false}, nil
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "unverified account",
			body: testLoginBody,
			mockFindById: func(ctx context.Context, id string) (prospect.Prospect, error) {
				return prospect.Prospect{ID: id, Status: prospect.StatusPending}, nil
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "credential service failure",
			body: testLoginBody,
			mockVerifyCredential: func(ctx context.Context, in *credentialv1.VerifyCredentialRequest, opts ...grpc.CallOption) (*credentialv1.VerifyCredentialResponse, error) {
				return nil, status.Error(codes.Unavailable, "connection refused")
			},
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockProspectRepository{
				FindByIdFunc: tt.mockFindById,
			}

			fillProspectDefaults(mockRepo)

			// init credential service mock
			mockCredentail := &MockCredentialServiceClient{
				VerifyCredentialFunc: tt.mockVerifyCredential,
			}

			fillCredentialDefaults(mockCredentail)

			// init email service mock
			mockEmail := &MockEmailService{}

			fillEmailDefaults(mockEmail)

			// inject mock to service
			rh := NewRegisterHandler(
				mockRepo,
				mockCredentail,
				mockEmail,
			)

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			rh.HandleLogin(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("http status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var resp LoginResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if resp.Status != prospect.StatusActive {
					t.Errorf("status = %q, want %q", resp.Status, prospect.StatusActive)
				}
			}
		})
	}
}

func TestHandleVerifyEmail(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockRepo := prospectmocks.NewMockRepository(ctrl)
	mockCredential := credentialmocks.NewMockCredentialServiceClient(ctrl)
	mockEmail := emailmocks.NewMockEmailService(ctrl)

	rh := NewRegisterHandler(
		mockRepo,
		mockCredential,
		mockEmail,
	)

	// test repo func for email verification
	mockRepo.EXPECT().
		FindByEmail(gomock.Any(), "loine@example.com").
		Return(testPendingProspect, nil)

	mockRepo.EXPECT().
		Active(gomock.Any(), "prospect-id-2").
		Return(nil)

	// test verify email flow
	req := httptest.NewRequest(http.MethodPost, "/verify-email", strings.NewReader(testVerifyEmailBody))
	rec := httptest.NewRecorder()
	rh.HandleVerifyEmail(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("http status = %d, want %d", rec.Code, http.StatusOK)
	}
}
